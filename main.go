package main

import (
	"cablenet/transcode"
	"cablenet/utilities"
	"fmt"
	"time"
)

type StreamState struct {
	metadata			utilities.StreamList
	transcoder		*transcode.Feed
}

var terminateTimeoutMil time.Duration = time.Duration(2000) * time.Millisecond
var ffparams utilities.FFoptions				// default ffmpeg options

var streamdata []StreamState						// metadata and reference to the process transcoding an active feed

func loadStreams() {
// load feeds
	var streamListParsed []utilities.StreamList
	utilities.LoadJsonFile("feeds.json", &streamListParsed)
	fmt.Println("\nStream Metadata: \n", utilities.GetStreamListInfo(streamListParsed))

	// add feed to managed list of feeds
	for _, f := range streamListParsed {
		sd := StreamState{
			metadata: f,
			transcoder: nil,
		}
		streamdata = append(streamdata, sd)
	}
}

func main() {
	loadStreams()

	// load ffmpeg default params
	utilities.LoadJsonFile("ffmpegconf.json", &ffparams)

	// testing
	ffstring, err := utilities.BuildFFparams(streamdata[0].metadata, ffparams)
	if err != nil {
		fmt.Println(err)
	}

	feed, err := transcode.StartFeed(ffstring...)
	if err != nil {
		fmt.Println("Error: ", err)
	}
	streamdata[0].transcoder = feed

	// input loop for testing
	var cmd string;
	for {
		fmt.Println("\nCablenet // Press x to exit")
		fmt.Print("Enter command: ")
		fmt.Scan(&cmd)
		fmt.Println("Executing ", cmd)

		if cmd == "x" {
			// stop feed
			stopErr := transcode.StopFeed(streamdata[0].transcoder, terminateTimeoutMil)
			fmt.Println("Stopped feed with code ", stopErr)

			fmt.Println("Goodbye!")
			break
		}
	}
	
}

