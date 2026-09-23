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

	/* Exit status of the process itself */
	done		chan error

	/* Additional status indicator after the process exits; channel closes when termination is complete */
	stopped chan struct{}

	/* Flag indicating that the feed should be stopped normally */
	voluntaryStop chan bool

	/* Logger */
	logger *lumberjack.Logger
}

const logPath string = "./log/ffmpeg/"
const terminateTimeoutMil time.Duration = time.Duration(2000) * time.Millisecond

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
	
	cmd := exec.Command("ffmpeg", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

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

	// Additional communication channels
	stopped := make (chan struct{}, 1)
	voluntaryStop := make (chan bool, 1)

	feed := Feed{
		proc: cmd,
		done: done,
		stopped: stopped,
		voluntaryStop: voluntaryStop,
		logger: fflogger,
	}

	// start monitor goroutine
	go streamMonitor(&feed)

	return &feed, nil
}

// Stream monitor process - restart feeds if stopped unexpectedly
func streamMonitor(f *Feed) {
	var procStatus error
	
	select {
	case <- f.done:
		// restart logic
		log.Printf("Restarting feed...")
		close(f.stopped)
		
	case <- f.voluntaryStop:
		// Stop feed
		sigtermErr := f.proc.Process.Signal(syscall.SIGTERM)
		if sigtermErr != nil {
			log.Println("WARNING: Error on SIGTERM attempt, ", sigtermErr)
		}

		select {
		case procStatus = <- f.done:
			log.Println("Feed stopped with code ", procStatus)
			close(f.stopped)
		case <- time.After(terminateTimeoutMil):
			sigkillErr := f.proc.Process.Kill()

			if sigkillErr != nil {
				log.Println("WARNING: Error on SIGKILL attempt, ", sigkillErr)
			} else {
				log.Println("Warning: Unresponsive transcode process killed")
			}

			<- f.done
			close(f.stopped)
		}
	}
}

func StopFeed(feed *Feed) error {
	// set voluntary stop flag
	log.Println("Sending stop signal...")
	feed.voluntaryStop <- true

	<- feed.stopped
	if err := feed.logger.Close(); err != nil {
		log.Println("Error closing feed logger: ", err)
	}

	return nil
}
