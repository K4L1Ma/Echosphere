package main

import (
	"context"
	"errors"
	"github.com/chiguirez/snout/v2"
	"github.com/k4l1ma/EchoSphere/build/client"
)

func main() {
	kernel := snout.Kernel{
		RunE: client.Run,
	}

	kernelBootstrap := kernel.Bootstrap(
		context.Background(),
		new(client.Config),
	)

	if err := kernelBootstrap.Initialize(); err != nil {
		if !errors.Is(err, context.Canceled) {
			panic(err)
		}
	}
}
