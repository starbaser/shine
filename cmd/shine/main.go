package main

import (
	"fmt"
	"os"
)

const version = "0.2.0"

func usage() {
	fmt.Print(renderHelp(usageHelp))
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	command := os.Args[1]

	// Handle flags
	switch command {
	case "-h", "--help":
		usage()
		return
	case "-v", "--version":
		fmt.Printf("shine v%s\n", version)
		return
	}

	// Execute command
	var err error
	switch command {
	case "start":
		err = cmdStart()

	case "stop":
		err = cmdStop()

	case "reload":
		err = cmdReload()

	case "status":
		err = cmdStatus()

	case "logs":
		panelID := ""
		if len(os.Args) > 2 {
			panelID = os.Args[2]
		}
		err = cmdLogs(panelID)

	default:
		Error(fmt.Sprintf("Unknown command: %s", command))
		fmt.Println()
		usage()
		os.Exit(1)
	}

	if err != nil {
		Error(err.Error())
		os.Exit(1)
	}
}
