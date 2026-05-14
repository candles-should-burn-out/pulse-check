package internal

import "time"

type Config struct {
	Addr            string
	ServiceName     string
	ShutdownTimeout time.Duration
	Auth            AuthConfig
}

type AuthConfig struct {
	Issuer       string
	JWKSURL      string
	Audience     string
	RequiredRole string
}
