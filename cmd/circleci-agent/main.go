package main

import (
	"fmt"
	"os"

	agent "github.com/matsune/circleci-agent"
)

func usage() {
	fmt.Print(`Usage:
	circleci-agent TARGET [OPTIONS]

Application Options:
  -v, --version      Show version

Help Options:
  -h, --help         Show this help message
`)
}

const version = "1.0"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	for _, a := range os.Args {
		if a == "-v" || a == "--version" {
			fmt.Printf("circleci-agent version %s\n", version)
			os.Exit(0)
		}
		if a == "-h" || a == "--help" {
			usage()
			os.Exit(0)
		}
	}

	target := os.Args[1]

	if err := agent.Run(target); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
