package main

import (
	"context"
	"errors"
	"github.com/chiguirez/snout/v3"
	"github.com/k4l1ma/EchoSphere/build/client"
)

func main() {
	kernel := snout.Kernel[client.Config]{RunE: client.Run}

	if err := kernel.Bootstrap(context.Background()).Initialize(); err != nil {
		if !errors.Is(err, context.Canceled) {
			panic(err)
		}
	}
}
