package main

import (
	"albanog/timer/timer"

	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	timeout := time.Second * 10
	m := timer.New(timeout)
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Printf("There's been an error: %v", err)
		os.Exit(1)
	}
}
