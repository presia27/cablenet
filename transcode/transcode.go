package transcode

import (
	"fmt"
	"log"
	"os/exec"
	"syscall"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

type Feed struct {
	/* Reference to the running process */
	proc		*exec.Cmd

	/* Process exit status */
	done		chan error

	/* Logger */
	logger *lumberjack.Logger
}

const logPath string = "./log/ffmpeg/"

// streamid - provide a unique id number to identify the stream (mostly for logging purposes).
// args - ffmpeg transcoding arguments.
func StartFeed (streamid int, args ...string) (*Feed, error) {
	logName := fmt.Sprintf("transcode_stream%d.log", streamid)
	fflogger := &lumberjack.Logger{
		Filename:		logPath + logName,
		MaxSize:		25,		// mbps before rotating
		MaxBackups:	10,		// number of old files to keep
		MaxAge:			14,		// days
		Compress:		true,	// gzip rotated files
	}
	defer fflogger.Close()
	
	cmd := exec.Command("ffmpeg", args...)

	// Direct ffmpeg progress output
	cmd.Stderr = fflogger

	// Start process
	err := cmd.Start()

	if err != nil {
		log.Println(err)
		fflogger.Close()
		return nil, err
	}
	log.Println("Started ffmpeg process")

	// Claude: give Wait() to a goroutine
	done := make (chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	feed := Feed{
		proc: cmd,
		done: done,
		logger: fflogger,
	}

	return &feed, nil
}

func StopFeed(feed *Feed, timeout time.Duration) error {
	feed.proc.Process.Signal(syscall.SIGTERM)

	var waitErr error
	select {
	case waitErr = <-feed.done:
		// receive status from channel and store
	case <-time.After(timeout): // force kill after timeout
		feed.proc.Process.Kill()
		log.Println("Warning: Unresponsive transcode process killed")
		waitErr = <-feed.done
	}

	if err := feed.logger.Close(); err != nil {
		log.Println("Error closing feed logger: ", err)
	}

	return waitErr
}
