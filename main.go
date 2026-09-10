package main

import (
	"cablenet/transcode"
	"fmt"
	"time"
)

var terminateTimeoutMil time.Duration = time.Duration(2000) * time.Millisecond

func main() {
	// temporary encode string
	encodeString := []string{
		"-y",
		"-i", "./KONG-HD_Harborstone_Junior_Achievement_emailquality.mp4",
		"-c:v", "libx264",
		"-profile:v", "high",
		"-pix_fmt", "yuv420p",
		"-colorspace", "bt709",
		"-color_primaries", "bt709",
		"-color_trc", "bt709",
		"-color_range", "tv",
		"-c:a", "aac",
		"-b:a", "384k",
		"testing.mp4",
	}

	fmt.Println(encodeString)

	feed, _ := transcode.StartFeed(encodeString...)

	// input loop for testing
	var cmd string;
	for {
		fmt.Println("\nCablenet // Press q to quit")
		fmt.Print("Enter command: ")
		fmt.Scan(&cmd)
		fmt.Println("Executing ", cmd)

		if cmd == "q" {
			// stop feed
			stopErr := transcode.StopFeed(feed, terminateTimeoutMil)
			fmt.Println("Stopped feed with code ", stopErr)

			fmt.Println("Goodbye!")
			break
		}
	}
	
}

