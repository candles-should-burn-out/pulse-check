package internal

import "time"

type Config struct {
	Addr            string
	ShutdownTimeout time.Duration
}
