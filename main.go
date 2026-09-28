package main

import (
	"cablenet/streamcontrol"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	sc := streamcontrol.Default()

	sc.LoadStreams("feeds.json")
	sc.Autostart()

	fmt.Println("\n/// Cablenet ///")

	sig := make(chan os.Signal, 2)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)

	<-sig
	// Shutdown procedure
	fmt.Println("Shutting down streams...")
	time.Sleep(200 * time.Millisecond) // sleep to avoid ffmpeg immediate exit from multiple interrupts

	shutdownDone := make(chan struct{})
	go func() {
		sc.StopAllStreams()
		close(shutdownDone)
	}()

	select {
	case <-shutdownDone:
		fmt.Println("Goodbye!")
	case <-sig:
		fmt.Println("Immediate exit requested from second signal. Check for running ffmpeg processes manually as they may have been abandoned by this abort.")
		os.Exit(1)
	}

}
