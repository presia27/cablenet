package streamcontrol

import (
	"cablenet/transcode"
	"cablenet/utilities"
	"errors"
	"fmt"
	"log"
	"sync"
)

const ffconfigfile string = "ffmpegconf.json"

// Valid status types for each stream
type StreamStatus string
const (
	Stopped StreamStatus = "stopped"
	Started	StreamStatus = "started"
	Failed	StreamStatus = "failed"
	Unknown	StreamStatus = "unknown"
)

// Information on stream metadata and the transcoding process
// A nil transcoder means that the feed is not active (no transcoding process running)
type StreamState struct {
	mu         sync.Mutex
	metadata   utilities.StreamList
	transcoder *transcode.Feed
	exitFromFailure		bool
}

// Parameters for the stream controller
// streamdata - data structure for streams (StreamState)
// ffparams - parameters for ffmpeg transcoding
type StreamControl struct {
	ffparams		utilities.FFoptions
	streamdata		[]*StreamState
}

// Initialize defaults for the stream controller
func Default() *StreamControl {
	var streamdata []*StreamState

	var ffparams utilities.FFoptions
	utilities.LoadJsonFile(ffconfigfile, &ffparams)

	sc := StreamControl{
		streamdata: streamdata,
		ffparams: ffparams,
	}

	return &sc
}

// Load stream metadata from file (filename)
func (sc *StreamControl) LoadStreams(filename string) {
	// load feeds
	var streamListParsed []utilities.StreamList
	utilities.LoadJsonFile(filename, &streamListParsed)
	//fmt.Println("\nStream Metadata: \n", utilities.GetStreamListInfo(streamListParsed))

	// add feed to managed list of feeds
	for _, f := range streamListParsed {
		sd := StreamState{
			metadata:   f,
			transcoder: nil,
		}
		sc.streamdata = append(sc.streamdata, &sd)
	}
}

// Start a feed by beginning the transcoding process and attaching the transcode feed struct
// to the StreamState
func (sc *StreamControl) startStream(s *StreamState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.exitFromFailure = false

	if s.transcoder != nil {
		return fmt.Errorf("Stream %d %s already started. No action taken.", s.metadata.ChannelNum, s.metadata.Name)
	}

	if !s.metadata.Enabled {
		return fmt.Errorf("Stream %d %s is disabled and cannot be started.", s.metadata.ChannelNum, s.metadata.Name)
	}

	// build ffmpeg command
	ffstring, err := utilities.BuildFFparams(s.metadata, sc.ffparams)
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
		err := <-chanFail

		s.mu.Lock()
		log.Printf("Failed to load/reload channel %d, status: %s", s.metadata.ChannelNum, err)
		s.transcoder = nil
		s.exitFromFailure = true
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
	if err != nil {
		return err
	}

	s.transcoder = nil

	return nil
}

// Stop all running streams
func (sc *StreamControl) StopAllStreams() {
	for _, f := range sc.streamdata {
		err := stopStream(f)
		if err != nil {
			log.Println(err)
		}
	}
}

// Start streams marked with the autostart flag
func (sc *StreamControl) Autostart() {
	for i := 0; i < len(sc.streamdata); i++ {
		if sc.streamdata[i].metadata.Enabled && sc.streamdata[i].metadata.AutostartEncode {
			err := sc.startStream(sc.streamdata[i])
			if err != nil {
				log.Println("Error starting stream: ", err)
			}
		}
	}
}

// Print the ffmpeg command that will be used on enabled streams
func (sc *StreamControl) PrintFFparams() {
	for i := 0; i < len(sc.streamdata); i++ {
		if sc.streamdata[i].metadata.Enabled {
			fmt.Println(utilities.BuildFFparams(sc.streamdata[i].metadata, sc.ffparams))
		}
	}
}

func (sc *StreamControl) GetMetadataByChannel(ch int) (utilities.StreamList, error) {
	for i := 0; i < len(sc.streamdata); i++ {
		if sc.streamdata[i].metadata.ChannelNum == ch {
			return sc.streamdata[i].metadata, nil
		}
	}
	return utilities.StreamList{}, errors.New("Unable to find channel")
}

func (sc *StreamControl) GetStreamStatus(ch int) (StreamStatus, error) {
	for i := 0; i < len(sc.streamdata); i++ {
		if sc.streamdata[i].metadata.ChannelNum == ch {
			if sc.streamdata[i].transcoder != nil {
				return Started, nil
			} else {
				if sc.streamdata[i].exitFromFailure {
					return Failed, nil
				} else {
					return Stopped, nil
				}
			}
		}
	}
	return Unknown, errors.New("Unable to find channel")
}

func (sc *StreamControl) GetAllMetadata() ([]utilities.StreamList) {
	var metaList []utilities.StreamList

	for i := 0; i < len(sc.streamdata); i++ {
		metaList = append(metaList, sc.streamdata[i].metadata)
	}

	return metaList
}
