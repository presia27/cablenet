package main

import (
	"cablenet/transcode"
	"cablenet/utilities"
	"fmt"
	"time"
)

type StreamState struct {
	metadata			utilities.StreamList
	transcoder		transcode.Feed
}

var terminateTimeoutMil time.Duration = time.Duration(2000) * time.Millisecond
var streamlist []utilities.StreamList		// list of feeds and their metadata
var ffparams utilities.FFoptions				// default ffmpeg options

var streamdata []StreamState						// metadata and reference to the process transcoding an active feed

func main() {
	// load feeds
	utilities.LoadJsonFile("feeds.json", &streamlist)
	streamInfo := utilities.GetStreamListInfo(streamlist)
	fmt.Println("\nStream Metadata: \n", streamInfo)

	// load ffmpeg default params
	utilities.LoadJsonFile("ffmpegconf.json", &ffparams)

	ffstring, err := utilities.BuildFFparams(streamlist[0], ffparams)
	if err != nil {
		fmt.Println(err)
	}

	feed, err := transcode.StartFeed(ffstring...)
	if err != nil {
		fmt.Println("Error: ", err)
	}
	var feedStreamState StreamState;
	feedStreamState = StreamState{
		metadata: streamlist[0],
		transcoder: *feed,
	}

	// Add to list of managed streams
	streamdata = append(streamdata, feedStreamState)
	fmt.Print(streamdata)

	//feed, _ := transcode.StartFeed(encodeString...)

	// input loop for testing
	var cmd string;
	for {
		fmt.Println("\nCablenet // Press x to exit")
		fmt.Print("Enter command: ")
		fmt.Scan(&cmd)
		fmt.Println("Executing ", cmd)

		if cmd == "x" {
			// stop feed
			stopErr := transcode.StopFeed(&streamdata[0].transcoder, terminateTimeoutMil)
			fmt.Println("Stopped feed with code ", stopErr)

			fmt.Println("Goodbye!")
			break
		}
	}
	
}

