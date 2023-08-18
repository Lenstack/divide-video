package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	inputPath := "video.mp4"
	outputPath := "output"
	chunkDuration := 180 // 3 minutes
	probePath := "ffmpeg-master-latest-win64-gpl/bin/"

	err := splitVideo(inputPath, outputPath, probePath, chunkDuration)
	if err != nil {
		log.Fatal(err)
	}
}

func splitVideo(inputPath, outputPath, probePath string, chunkDuration int) error {
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
	// Split the video into chunks
	for i := 0; i < numChunks; i++ {
		startTime := i * chunkDuration
		outputFilename := fmt.Sprintf("chunk_%d.mp4", i)
		outputPath := filepath.Join(outputPath, outputFilename)

		cmd := exec.Command(probePath+"ffmpeg.exe", "-ss", fmt.Sprintf("%d", startTime), "-i", inputPath, "-t", fmt.Sprintf("%d", chunkDuration), "-c", "copy", outputPath)
		err := cmd.Run()
		if err != nil {
			return err
		}
		fmt.Printf("Chunk %d/%d done\n", i+1, numChunks)
	}
	fmt.Printf("Done splitting video into %d chunks\n", numChunks)
	return nil
}
