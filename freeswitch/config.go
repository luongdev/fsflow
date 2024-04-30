package freeswitch

import "time"

type Config struct {
	FsHost           string        `yaml:"fsHost"`
	FsPort           uint16        `yaml:"fsPort"`
	FsPassword       string        `yaml:"fsPassword"`
	Timeout          time.Duration `yaml:"timeout"`
	ServerListenPort uint16        `yaml:"serverListenPort"`
}
