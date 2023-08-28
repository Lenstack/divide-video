package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	inputPath        = "videos/DarkGathering_8.mp4"
	outputPath       = "output"
	mutedInputPath   = ""
	chunkDuration    = 180 // 3 minutes
	probePath        = "ffmpeg-master-latest-win64-gpl/bin/"
	timeRangesToMute = []TimeRange{
		{StartTime: "00:00:10", EndTime: "00:01:00"},
	}
)

type TimeRange struct {
	StartTime string
	EndTime   string
}

func main() {
	// Split the video into chunks
	err := splitAndMuteVideo(inputPath, outputPath, probePath, chunkDuration, timeRangesToMute)
	if err != nil {
		log.Fatal(err)
	}
}

func splitAndMuteVideo(inputPath, outputPath, probePath string, chunkDuration int, timeRangesToMute []TimeRange) error {
	// Create output directory if it doesn't exist
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		err := os.Mkdir(outputPath, 0755)
		if err != nil {
			return err
		}
	}

	// Get video duration using ffprobe
	cmd := exec.Command(probePath+"ffprobe.exe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", inputPath)
	durationOutput, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	// Convert duration output to float64
	durationStr := strings.TrimSpace(string(durationOutput))
	var duration float64
	_, err = fmt.Sscanf(durationStr, "%f", &duration)
	if err != nil {
		return err
	}
	fmt.Printf("Video duration: %f\n", duration)

	// Calculate the number of chunks
	numChunks := int(duration) / chunkDuration
	numChunks++
	fmt.Printf("Splitting video into %d chunks\n", numChunks)

	// Validate if the video needs to be muted
	if len(timeRangesToMute) > 0 {
		// Mute the video in the specified time ranges
		fmt.Printf("Muting video in %d time ranges\n", len(timeRangesToMute))
		for _, timeRange := range timeRangesToMute {
			// Convert time range to seconds
			startTime, err := convertDurationToSeconds(timeRange.StartTime)
			if err != nil {
				log.Fatal(err)
			}
			endTime, err := convertDurationToSeconds(timeRange.EndTime)
			if err != nil {
				log.Fatal(err)
			}

			cmd := exec.Command(probePath+"ffmpeg.exe", "-i", inputPath, "-af", fmt.Sprintf("volume=enable='between(t,%d,%d)':volume=0", startTime, endTime), "-c:v", "copy", "-c:a", "aac", "-strict", "-2", filepath.Join(outputPath, "muted_"+filepath.Base(inputPath)))
			err = cmd.Run()
			if err != nil {
				return err
			}
			fmt.Printf("Muted video from %d to %d\n", timeRange.StartTime, timeRange.EndTime)
		}

		// Set the muted input path
		mutedInputPath = filepath.Join(outputPath, "muted_"+filepath.Base(inputPath))

		fmt.Printf("Muting video in %d time ranges\n", len(timeRangesToMute))
	} else {
		mutedInputPath = inputPath
		fmt.Println("No time ranges to mute")
	}

	// Split the video into chunks
	for i := 0; i < numChunks; i++ {
		startTime := i * chunkDuration
		fileName := filepath.Base(inputPath)
		outputFilename := fmt.Sprintf("%s_%d.mp4", strings.TrimSuffix(fileName, filepath.Ext(fileName)), i+1)
		outputPath := filepath.Join(outputPath, outputFilename)

		cmd := exec.Command(probePath+"ffmpeg.exe", "-ss", fmt.Sprintf("%d", startTime), "-i", mutedInputPath, "-t", fmt.Sprintf("%d", chunkDuration), "-c", "copy", outputPath)
		err := cmd.Run()
		if err != nil {
			return err
		}
		fmt.Printf("Chunk %d/%d done\n", i+1, numChunks)
	}

	// Delete the muted video file if it exists
	if _, err := os.Stat(filepath.Join(outputPath, "muted_"+filepath.Base(inputPath))); !os.IsNotExist(err) {
		err := os.Remove(filepath.Join(outputPath, "muted_"+filepath.Base(inputPath)))
		if err != nil {
			return err
		}
	}

	fmt.Println("Deleted muted video")
	fmt.Printf("Done splitting video into %d chunks\n", numChunks)
	return nil
}

func convertDurationToSeconds(durationStr string) (int, error) {
	parts := strings.Split(durationStr, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid duration format")
	}

	totalSeconds := 0
	multipliers := []int{3600, 60, 1}

	for i, part := range parts {
		value, err := strconv.Atoi(part)
		if err != nil {
			return 0, err
		}
		totalSeconds += value * multipliers[i]
	}

	return totalSeconds, nil
}
