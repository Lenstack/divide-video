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

type TimeRange struct {
	StartTime string
	EndTime   string
}

type VideoDivider struct {
	inputVideoPath   string
	outputVideoPath  string
	mutedVideoPath   string
	chunkDuration    string
	ffPath           string
	timeRangesToMute []TimeRange
}

func main() {
	videoDivider := VideoDivider{
		inputVideoPath:  "videos/DarkGathering_8.mp4",
		outputVideoPath: "output",
		mutedVideoPath:  "",
		chunkDuration:   "00:03:00",
		ffPath:          "ffmpeg-master-latest-win64-gpl/bin",
		timeRangesToMute: []TimeRange{
			{StartTime: "00:01:53", EndTime: "00:03:22"},
			{StartTime: "00:21:51", EndTime: "00:23:20"},
		},
	}
	videoDivider.ProcessVideo()
}

func (v *VideoDivider) ProcessVideo() {
	v.MuteVideo()
	v.DivideVideo()
}

func (v *VideoDivider) DivideVideo() {
	// Get video duration
	duration, err := v.GetVideoDuration()
	if err != nil {
		log.Fatalf("Error getting video duration: %v", err)
	}
	log.Printf("Video duration: %v", duration)

	// Create output folder
	err = v.CreateOutputFolder()
	if err != nil {
		log.Fatalf("Error creating output folder: %v", err)
	}

	// Convert chunk duration to seconds
	chunkDuration, err := v.ConvertDurationToSeconds(v.chunkDuration)
	if err != nil {
		log.Fatalf("Error converting chunk duration: %v", err)
	}
	log.Printf("Chunk duration: %v", chunkDuration)

	// Calculate number of chunks
	numChunks := int(duration) / chunkDuration
	log.Printf("Number of chunks: %v", numChunks)

	// Divide video into chunks and save to output folder
	for i := 0; i < numChunks; i++ {
		// Calculate start time and end time
		startTime := i * chunkDuration

		// Create file name for chunk
		fileName := fmt.Sprintf("%s_%d.mp4", strings.TrimSuffix(filepath.Base(v.inputVideoPath), filepath.Ext(v.inputVideoPath)), i+1)

		// Execute ffmpeg command to divide video
		cmd := v.ExecuteFFCommand("ffmpeg.exe", []string{
			"-ss", fmt.Sprintf("%d", startTime),
			"-i", v.mutedVideoPath,
			"-t", fmt.Sprintf("%d", chunkDuration),
			"-c", "copy",
			filepath.Join(v.outputVideoPath, fileName),
		})

		// Execute command and check for errors
		err = cmd.Run()
		if err != nil {
			log.Fatalf("Error dividing video: %v", err)
		}

		log.Printf("Chunk %d/%d done\n", i+1, numChunks)
	}

	// Delete muted video from output folder if exists
	err = v.DeleteMutedVideo()
	if err != nil {
		log.Fatalf("Error deleting muted video: %v", err)
	}
	log.Printf("Deleted muted video from output folder")
}

func (v *VideoDivider) MuteVideo() {
	log.Printf("Muting video in %d time ranges", len(v.timeRangesToMute))
	for _, timeRange := range v.timeRangesToMute {
		startTime, _ := v.ConvertDurationToSeconds(timeRange.StartTime)
		endTime, _ := v.ConvertDurationToSeconds(timeRange.EndTime)

		cmd := v.ExecuteFFCommand("ffmpeg.exe", []string{
			"-i", v.inputVideoPath,
			"-af", fmt.Sprintf("volume=enable='between(t,%d,%d)':volume=0", startTime, endTime),
			"-c:v", "copy",
			"-c:a", "aac",
			"-strict", "-2",
			filepath.Join(v.outputVideoPath, "muted_"+filepath.Base(v.inputVideoPath))})

		// Execute command and check for errors
		err := cmd.Run()
		if err != nil {
			log.Fatalf("Error muting video: %v", err)
		}

		log.Printf("Muted video from %v to %v", timeRange.StartTime, timeRange.EndTime)
	}

	v.mutedVideoPath = filepath.Join(v.outputVideoPath, "muted_"+filepath.Base(v.inputVideoPath))
}

func (v *VideoDivider) ConvertDurationToSeconds(duration string) (int, error) {
	parts := strings.Split(duration, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid duration format")
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}

	seconds, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, err
	}

	totalSeconds := hours*3600 + minutes*60 + seconds
	return totalSeconds, nil
}

func (v *VideoDivider) ExecuteFFCommand(executable string, args []string) *exec.Cmd {
	// Create command to execute
	return exec.Command(filepath.Join(v.ffPath, executable), args...)
}

func (v *VideoDivider) GetVideoDuration() (float64, error) {
	// Execute ffprobe command to get duration
	cmd := v.ExecuteFFCommand("ffprobe.exe", []string{"-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", v.inputVideoPath})

	// Get duration output from command
	durationOutput, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error getting video duration: %v", err)
	}

	// Convert duration output to float64
	durationStr := strings.TrimSpace(string(durationOutput))
	var duration float64
	_, err = fmt.Sscanf(durationStr, "%f", &duration)
	if err != nil {
		log.Fatalf("Error converting duration to float64: %v", err)
	}

	return duration, nil
}

func (v *VideoDivider) CreateOutputFolder() error {
	// Check if output folder exists and create if not exists
	if _, err := os.Stat(v.outputVideoPath); os.IsNotExist(err) {
		err := os.Mkdir(v.outputVideoPath, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func (v *VideoDivider) CreateMutedVideoFolder() error {
	// Check if output folder exists and create if not exists
	if _, err := os.Stat(v.mutedVideoPath); os.IsNotExist(err) {
		err := os.Mkdir(v.mutedVideoPath, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func (v *VideoDivider) DeleteMutedVideoFolder() error {
	// Check if output folder exists and delete if exists
	if _, err := os.Stat(v.mutedVideoPath); os.IsExist(err) {
		err := os.RemoveAll(v.mutedVideoPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func (v *VideoDivider) DeleteMutedVideo() error {
	// Check if the muted video file exists
	if _, err := os.Stat(v.mutedVideoPath); err == nil {
		// Delete the muted video file
		err := os.Remove(v.mutedVideoPath)
		if err != nil {
			return err
		}
	}
	return nil
}
