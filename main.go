package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	inputPath        = "videos/DarkGathering_8.mp4"
	outputPath       = "output"
	mutedInputPath   string
	chunkDuration    = 180 // 3 minutes
	probePath        = "ffmpeg-master-latest-win64-gpl/bin/"
	timeRangesToMute = []TimeRange{
		{StartTime: 130, EndTime: 200},
		{StartTime: 1320, EndTime: 1380},
	}
)

type TimeRange struct {
	StartTime int
	EndTime   int
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

	if len(timeRangesToMute) > 0 {
		// Mute the video in the specified time ranges
		fmt.Printf("Muting video in %d time ranges\n", len(timeRangesToMute))
		for _, timeRange := range timeRangesToMute {
			cmd := exec.Command(probePath+"ffmpeg.exe", "-i", inputPath, "-af", fmt.Sprintf("volume=enable='between(t,%d,%d)':volume=0", timeRange.StartTime, timeRange.EndTime), "-c:v", "copy", "-c:a", "aac", "-strict", "-2", filepath.Join(outputPath, "muted_"+filepath.Base(inputPath)))
			err := cmd.Run()
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

	// Delete the muted video
	err = os.Remove(mutedInputPath)
	if err != nil {
		return err
	}
	fmt.Printf("Done splitting video into %d chunks\n", numChunks)
	return nil
}
