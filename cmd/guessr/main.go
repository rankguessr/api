package main

import (
	"context"
	"log"

	"github.com/rankguessr/api/internal/cli"
)

func main() {
	ctx := context.Background()
	if err := cli.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
