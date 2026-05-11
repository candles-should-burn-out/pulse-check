package main

import "time"

type Config struct {
	Addr            string
	ShutdownTimeout time.Duration
}
