package transcode

import (
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
}

func StartFeed (args ...string) (*Feed, error) {
	fflogger := &lumberjack.Logger{
		Filename:		"./log/ffmpeg/transcode.log",
		MaxSize:		50,		// mbps before rotating
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
		log.Fatal(err)
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
	}

	return &feed, nil
}

func StopFeed(feed *Feed, timeout time.Duration) error {
	feed.proc.Process.Signal(syscall.SIGTERM)

	select {
	case err := <-feed.done:
		return err
	case <-time.After(timeout): // force kill after timeout
		feed.proc.Process.Kill()
		return <-feed.done
	}
}
