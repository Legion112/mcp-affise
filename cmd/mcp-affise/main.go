package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Legion112/mcp-affise/internal/app"
)

func main() {
	if err := app.Run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "mcp-affise: %v\n", err)
		os.Exit(1)
	}
}
