package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please specify 'prepare' or 'run' as an argument to this application.")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "prepare":
		runHistogram()
	case "run":
		runPerceptron()
	}
}
