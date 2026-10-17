package main

import (
	"fmt"
	"os"
)

func main() {
	config, err := NewConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "wakafetch:", err)
		fmt.Fprintln(os.Stderr, "Run 'wakafetch --help' for usage.")
		os.Exit(1)
	}

	stats, err := GetStats(config)
	if err != nil {
		fmt.Fprintln(os.Stderr, "wakafetch:", err)
		os.Exit(1)
	}

	err = Render(*config, *stats)
	if err != nil {
		fmt.Fprintln(os.Stderr, "wakafetch:", err)
		os.Exit(1)
	}
}
