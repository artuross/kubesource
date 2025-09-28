package main

import (
	"context"
	"os"

	"github.com/artuross/kubesource/internal/commands"
)

func main() {
	rootCmd := commands.NewKubesourceCommand()

	if err := rootCmd.Run(context.Background(), os.Args); err != nil {
		os.Exit(1)
	}
}
