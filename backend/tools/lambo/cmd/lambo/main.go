package main

import (
	"fmt"
	"os"

	"github.com/Edward-Jackie/lambo/backend/tools/lambo/internal/command"
)

func main() {
	err := command.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, "错误：", err)
		os.Exit(1)
	}
}
