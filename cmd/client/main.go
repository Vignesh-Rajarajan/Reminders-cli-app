package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/vickey290/go-modules/client"
)

var (
	backendURIFlag = flag.String("backednd", "http://localhost:8080", "Backend API URL")
	help           = flag.Bool("help", false, "use to display help texts")
)

func main() {
	flag.Parse()
	s := client.NewSwitch(*backendURIFlag)

	if *help || len(os.Args) == 1 {
		s.Help()
		return
	}

	err := s.Switch()

	if err != nil {
		fmt.Printf("cmd switch error: %v\n ", err)
		os.Exit(2)
	}
}
