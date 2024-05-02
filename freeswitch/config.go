package freeswitch

import "time"

type Config struct {
	Host     string        `yaml:"host"`
	Port     uint16        `yaml:"port"`
	Password string        `yaml:"password"`
	Timeout  time.Duration `yaml:"timeout"`
	ListenOn uint16        `yaml:"listen_on"`
}
