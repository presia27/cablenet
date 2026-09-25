package main

import (
	"cablenet/transcode"
	"cablenet/utilities"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Information on stream metadata and the transcoding process
// A nil transcoder means that the feed is not active (no transcoding process running)
type StreamState struct {
	mu 						sync.Mutex
	metadata			utilities.StreamList
	transcoder		*transcode.Feed
}

var ffparams utilities.FFoptions				// default ffmpeg options

var streamdata []*StreamState						// metadata and reference to the process transcoding an active feed

func loadStreams() {
// load feeds
	var streamListParsed []utilities.StreamList
	utilities.LoadJsonFile("feeds.json", &streamListParsed)
	//fmt.Println("\nStream Metadata: \n", utilities.GetStreamListInfo(streamListParsed))

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
	s.mu.Lock()
	defer s.mu.Unlock()

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

	feed, chanFail, err := transcode.StartFeed(s.metadata.ChannelNum, ffstring...)
	if err != nil {
		return err
	}

	// Set transcoder struct field to active feed
	s.transcoder = feed

	// automatically remove transcoder pointer if the ffmpeg fails to load
	go func() {
		err := <- chanFail

		s.mu.Lock()
		log.Printf("Failed to load/reload channel %d, status: %s", s.metadata.ChannelNum, err)
		s.transcoder = nil
		s.mu.Unlock()
	}()

	return nil
}

// Stop a running feed by stopping the transcoding process and removing the transcode feed struct
func stopStream(s *StreamState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.transcoder == nil {
		return fmt.Errorf("Stream %d %s is already stopped. No action taken.", s.metadata.ChannelNum, s.metadata.Name)
	}

	err := transcode.StopFeed(s.transcoder)
	if (err != nil) {
		return err
	}

	s.transcoder = nil

	return nil
}

func stopAllStreams() {
	for _, f := range streamdata {
		err := stopStream(f)
		if err != nil {
			log.Println(err)
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
				log.Println("Error starting stream: ", err)
			}
		}
	}

	fmt.Println("\n/// Cablenet ///")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)

	<- sig
	// Shutdown procedure
	fmt.Println("Shutting down streams...")
	time.Sleep(200 * time.Millisecond) // sleep to avoid ffmpeg immediate exit from multiple interrupts
	stopAllStreams()

	fmt.Println("Goodbye!")
	
}

