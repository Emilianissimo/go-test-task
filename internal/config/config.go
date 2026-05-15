package config

import "time"

type Config struct {
	DBURL            string
	RedisAddr        string
	RedisPass        string
	AppPort          string
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	IdleTimeout      time.Duration
	SkinportUrl      string
	SkinportAppID    string
	SkinportCurrency string
	SkinportCacheTTL time.Duration
	SkinportTimeout  time.Duration
}
