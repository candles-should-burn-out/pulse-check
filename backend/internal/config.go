package internal

import "time"

type Config struct {
	Addr            string
	ServiceName     string
	ShutdownTimeout time.Duration
}
