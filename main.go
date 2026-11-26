package main

import (
	"cablenet/config"
	"cablenet/servercontrol"
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
	// Load ENV
	cfg, err := config.Load(".env")
	if err != nil {
		log.Fatal("Configuration error: ", err)
	}

	// Stream Control System
	sc := streamcontrol.Default()

	sc.LoadStreams("feeds.json")
	sc.Autostart()

	// File Server
	fileServer := http.FileServer(http.Dir(hlsDir))

	httpFileServ := &http.Server{
		Addr:			fmt.Sprintf(":%d", cfg.FilePort),
		Handler:	fileServer,
		ReadTimeout:	10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("Starting file server on port %d\n", cfg.FilePort)
		err := httpFileServ.ListenAndServe()
		if err != nil {
			log.Println(err)
		}
	}()

	// Control Server
	apiServMux := servercontrol.NewMux(sc)

	ctlApiServ := &http.Server{
		Addr:				fmt.Sprintf(":%d", cfg.ControlPort),
		Handler: 		apiServMux,
		ReadTimeout:	10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("Starting API control server on port %d\n", cfg.ControlPort)
		err := ctlApiServ.ListenAndServe()
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
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := ctlApiServ.Shutdown(ctx); err != nil {
			log.Println("API control server forced to shutdown")
		}

		sc.StopAllStreams()

		ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel2()

		if err := httpFileServ.Shutdown(ctx2); err != nil {
			log.Println("File server forced to shutdown", err)
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
