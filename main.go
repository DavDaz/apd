package main

import (
	"fmt"
	"os"

	"apd/cmd"
)

func main() {
	if err := cmd.Execute(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
