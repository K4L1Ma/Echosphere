package server

type Config struct {
	Server  SrvCfg     `snout:"server"`
	SideCar SideCarCfg `snout:"sidecar"`
}

type SrvCfg struct {
	Port int `snout:"port" default:"8080"`
}
type SideCarCfg struct {
	Enabled bool `snout:"enabled" default:"true"`
	Port    int  `snout:"port" default:"9090"`
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
