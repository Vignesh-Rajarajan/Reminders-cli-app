package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("No commmands provided")
		os.Exit(2)
	}

	cmd := os.Args[1]

	switch cmd {
	case "greet":
		// msg := "REMINDERS CLI- CLI Basics"
		// if len(os.Args) > 2 {

		// 	f := strings.Split(os.Args[2], "=")

		// 	if len(f) == 2 && f[0] == "--msg" {
		// 		msg = f[1]
		// 	}
		// }
		// fmt.Printf("hello and welcome %s\n", msg)
		greetCmd := flag.NewFlagSet("greet", flag.ExitOnError)
		msgFlag := greetCmd.String("msg", "REMINDERS CLI- CLI Basics", "message for greet command")
		err := greetCmd.Parse(os.Args[2:])

		if err != nil {
			log.Fatal(err.Error())
		}
		fmt.Printf("hello and welcome %s\n", *msgFlag)

	case "help":
		fmt.Println("Some Help Message")
	default:
		fmt.Printf("unkown command provided %s\n", cmd)
	}
}
