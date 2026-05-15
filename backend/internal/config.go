package internal

import "time"

type Config struct {
	Addr            string
	ServiceName     string
	ShutdownTimeout time.Duration
	Auth            AuthConfig
	DatabaseURL     string
	StatusSet       StatusSetConfig
}

type AuthConfig struct {
	Issuer       string
	JWKSURL      string
	Audience     string
	RequiredRole string
}

type StatusSetConfig struct {
	MaxStatuses int
}
