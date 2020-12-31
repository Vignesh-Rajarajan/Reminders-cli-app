package server

import (
	"log"
	"os"
	"os/signal"
)

type Stopper interface {
	Stop() error
}

func ListenForSignals(signals []os.Signal, apps ...Stopper) {
	c := make(chan os.Signal, 1)

	signal.Notify(c, signals...)
	sig := <-c
	log.Printf("recieved shutdown signal %v\n", sig.String())

	var errs []error

	for _, app := range apps {
		err := app.Stop()
		if err != nil {
			errs = append(errs, err)
		}
	}

	var exitCode int

	for _, err := range errs {
		log.Printf("could not stop the server %v\n", err)
		exitCode = 1
	}
	os.Exit(exitCode)
}
