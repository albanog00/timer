package main

import (
	"albanog/timer/timer"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if len(os.Args) != 2 || os.Args[1] == "help" {
		printHelp()
		os.Exit(0)
	}

	timeout, err := time.ParseDuration(os.Args[1])
	if err != nil {
		fmt.Printf("invalid duration provided.\n")
		printHelp()
		os.Exit(1)
	}

	m := timer.New(timeout)
	program := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Printf("usage: %s <duration>.\n", os.Args[0])
	fmt.Printf("example: %s 1h10m15s # starts a timer of 1 hour 10 minutes and 15 seconds.\n", os.Args[0])
}
