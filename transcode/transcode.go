package transcode

import (
	"fmt"
	"log"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

type Feed struct {
	/* Reference to the running process */
	proc		*exec.Cmd

	/* Exit status of the process itself */
	done		chan error

	/* Error information if the feed fails after trying/retrying to start */
	failed	chan error

	/* Additional status indicator after the process exits; channel closes when termination is complete */
	stopped chan struct{}

	/* Flag indicating that the feed should be stopped normally */
	voluntaryStop chan bool

	/* Logger */
	logger *lumberjack.Logger

	/* MUTEX Lock */
	mu					sync.Mutex

	/* ==} Status information {== */
	
	/* Exit status information (of error type) */
	exitStatus	error

	/* Time that the process exited */
	exitTime		time.Time

	/* Number of attempted restarts in a fixed period */
	retryCount	int

	/* Timestamp of the last retry attempt */
	retryTime		time.Time
}

const logPath string = "./log/ffmpeg/"
const terminateTimeoutMil time.Duration = time.Duration(4000) * time.Millisecond
const retryMaxCount int = 6;
const retryInterval time.Duration = time.Duration(1) * time.Minute

func startFeedProc (fflogger *lumberjack.Logger, args ...string) (*exec.Cmd, error) {
	cmd := exec.Command("ffmpeg", args...)
	// Use a separate process group so that ffmpeg doesn't receive SIGINT before the code sends it
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// Direct ffmpeg progress output
	cmd.Stderr = fflogger

	// Start process
	err := cmd.Start()

	if err != nil {
		return nil, err
	}

	return cmd, nil
}

// streamid - provide a unique id number to identify the stream (mostly for logging purposes).
// args - ffmpeg transcoding arguments.
func StartFeed (streamid int, args ...string) (*Feed, chan error, error) {
	logName := fmt.Sprintf("transcode_stream%d.log", streamid)
	fflogger := &lumberjack.Logger{
		Filename:		logPath + logName,
		MaxSize:		25,		// mbps before rotating
		MaxBackups:	10,		// number of old files to keep
		MaxAge:			14,		// days
		Compress:		true,	// gzip rotated files
	}

	cmd, err := startFeedProc(fflogger, args...)
	if err != nil {
		log.Println(err)
	  fflogger.Close()
	  return nil, nil, err
	}
	log.Println("Started ffmpeg process")

	// Claude: give Wait() to a goroutine
	done := make (chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	// Additional communication channels
	failed := make (chan error, 1)
	stopped := make (chan struct{}, 1)
	voluntaryStop := make (chan bool, 1)

	feed := Feed{
		proc: cmd,
		done: done,
		failed: failed,
		stopped: stopped,
		voluntaryStop: voluntaryStop,
		logger: fflogger,
		retryCount: 0,
		retryTime: time.Now(),
	}

	// start monitor goroutine
	go streamMonitor(&feed)

	return &feed, failed, nil
}

// Stream monitor process - restart feeds if stopped unexpectedly
func streamMonitor(f *Feed) {
	var procStatus error
	
	select {
	case procStatus = <- f.done:
		time.Sleep(500 * time.Millisecond) // add delay between retries

		retryNum := updateRetryCount(f)
		if retryNum > retryMaxCount {
			f.failed <- procStatus
			setExitStatus(f, procStatus)
			close(f.stopped)
			if err := f.logger.Close(); err != nil {
				log.Println("Error closing feed logger: ", err)
			}

		} else {
			// restart logic
			log.Printf("Attempt %d // Restarting feed", retryNum)
			cmd, err := startFeedProc(f.logger, f.proc.Args[1:]...)
			if err != nil {
				log.Println("Error restarting feed:", err)
				// f.failed <- err
				// setExitStatus(f, err)
				// close(f.stopped)
				return
			}

			go func() {
				f.done <- cmd.Wait()
			}()

			f.proc = cmd

			// This instance of the streamMonitor will end shortly...
			// ...must restart the streamMonitor
			go streamMonitor(f)
		}
		
		
	case <- f.voluntaryStop:
		// Stop feed
		sigtermErr := f.proc.Process.Signal(syscall.SIGTERM)
		if sigtermErr != nil {
			log.Println("WARNING: Error on SIGTERM attempt, ", sigtermErr)
		}

		select {
		case procStatus = <- f.done:
			setExitStatus(f, procStatus)
			log.Println("Feed stopped with code ", procStatus)
			close(f.stopped)
		case <- time.After(terminateTimeoutMil):
			sigkillErr := f.proc.Process.Kill()

			if sigkillErr != nil {
				log.Println("WARNING: Error on SIGKILL attempt, ", sigkillErr)
			} else {
				log.Println("Warning: Unresponsive transcode process killed")
			}

			procStatus = <- f.done
			setExitStatus(f, procStatus)

			close(f.stopped)
		}
	}
}

func setExitStatus(f *Feed, e error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.exitStatus = e
	f.exitTime = time.Now()
}

func updateRetryCount(f *Feed) int {
	f.mu.Lock()
	defer f.mu.Unlock()

	// reset count if it's been longer than the specified retry interval
	if time.Since(f.retryTime) > retryInterval {
		f.retryCount = 0
	}

	f.retryCount = f.retryCount + 1
	f.retryTime = time.Now()

	return f.retryCount
}

func GetExitStatus(f *Feed) (err error, exitTime time.Time) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.exitStatus, f.exitTime
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
