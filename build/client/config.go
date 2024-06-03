package client

import (
	"time"
)

type Config struct {
	Target   string        `snout:"target" default:"localhost:8080"`
	DeadLine time.Duration `snout:"deadline" default:"30s"`
	SideCar  SideCar       `snout:"sidecar"`
}

type SideCar struct {
	Enabled bool `snout:"enabled" default:"true"`
	Port    int  `snout:"port" default:"9091"`
}

func (c Config) GetSideCar() struct {
	Enabled bool
	Port    int
} {
	return struct {
		Enabled bool
		Port    int
	}(c.SideCar)
}
