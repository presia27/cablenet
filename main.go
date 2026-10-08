package main

import (
	"cablenet/streamcontrol"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const hlsDir string = "./hls"

func main() {
	// Stream Control System
	sc := streamcontrol.Default()

	sc.LoadStreams("feeds.json")
	sc.Autostart()

	// File Server
	fileServer := http.FileServer(http.Dir(hlsDir))

	s := &http.Server{
		Addr:			":4000",
		Handler:	fileServer,
		ReadTimeout:	10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Println("Starting file server")
		err := s.ListenAndServe()
		if err != nil {
			log.Println(err)
		}
	}()

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

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.Shutdown(ctx); err != nil {
			log.Println("Server forced to shutdown", err)
		}

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
