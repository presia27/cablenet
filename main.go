package main

import (
	"cablenet/transcode"
	"cablenet/utilities"
	"fmt"
	"time"
)

// Information on stream metadata and the transcoding process
// A nil transcoder means that the feed is not active (no transcoding process running)
type StreamState struct {
	metadata			utilities.StreamList
	transcoder		*transcode.Feed
}

const terminateTimeoutMil time.Duration = time.Duration(2000) * time.Millisecond
var ffparams utilities.FFoptions				// default ffmpeg options

var streamdata []*StreamState						// metadata and reference to the process transcoding an active feed

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
		streamdata = append(streamdata, &sd)
	}
}

// Start a feed by beginning the transcoding process and attaching the transcode feed struct
// to the StreamState
func startStream(s *StreamState) error {
	if s.transcoder != nil {
		return fmt.Errorf("Stream %d %s already started. No action taken.", s.metadata.ChannelNum, s.metadata.Name)
	}

	if !s.metadata.Enabled {
		return fmt.Errorf("Stream %d %s is disabled and cannot be started.", s.metadata.ChannelNum, s.metadata.Name)
	}

	// build ffmpeg command
	ffstring, err := utilities.BuildFFparams(s.metadata, ffparams)
	if err != nil {
		return err
	}

	feed, err := transcode.StartFeed(ffstring...)
	if err != nil {
		return err
	}

	// Set transcoder struct field to active feed
	s.transcoder = feed

	return nil
}

// Stop a running feed by stopping the transcoding process and removing the transcode feed struct
func stopStream(s *StreamState) error {
	if s.transcoder == nil {
		return fmt.Errorf("Stream %d %s is already stopped. No action taken.", s.metadata.ChannelNum, s.metadata.Name)
	}

	err := transcode.StopFeed(s.transcoder, terminateTimeoutMil)
	fmt.Println("Stopped feed with code ", err)

	s.transcoder = nil

	return nil
}

func stopAllStreams() {
	for _, f := range streamdata {
		if f.transcoder != nil {
			err := stopStream(f)
			if err != nil {
				fmt.Println("Error while stopping feed: ", err)
			}
		}
	}
}

func main() {
	loadStreams()

	// load ffmpeg default params
	utilities.LoadJsonFile("ffmpegconf.json", &ffparams)

	for i := 0; i < len(streamdata); i++ {
		if streamdata[i].metadata.Enabled && streamdata[i].metadata.AutostartEncode {
			err := startStream(streamdata[i])
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	for _, x := range streamdata {
		fmt.Println("\n", x)
	}

	// input loop for testing
	var cmd string;
	for {
		fmt.Println("\nCablenet // Press x to exit")
		fmt.Print("Enter command: ")
		fmt.Scan(&cmd)
		fmt.Println("Executing ", cmd)

		if cmd == "x" {
			// stop feed
			stopAllStreams()

			fmt.Println("Goodbye!")
			break
		}
	}
	
}

