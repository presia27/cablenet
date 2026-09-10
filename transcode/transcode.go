package transcode

import (
	"log"
	"os/exec"
	"syscall"
	"time"
)

type Feed struct {
	/* Reference to the running process */
	proc		*exec.Cmd

	/* Process exit status */
	done		chan error
}

func StartFeed (args ...string) (*Feed, error) {
	cmd := exec.Command("ffmpeg", args...)

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
