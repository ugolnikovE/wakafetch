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

	println(config.Range)

	stats, err := GetStats(config)

	if err != nil {
		fmt.Fprintln(os.Stderr, "wakafetch:", err)
	} else {
		println("time: ", stats.TotalSeconds)
		println("top lang: ", stats.Languages[0].Name)
	}

}
