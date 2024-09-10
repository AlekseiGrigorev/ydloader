// Copyright 2024 Aleksei Grigorev
// https://aleksvgrig.com, https://github.com/AlekseiGrigorev, aleksvgrig@gmail.com.
// Package define interfaces, structures and functions for working with application configuration
package config

// Db define database configuration
type Db struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

// Http define http configuration
type Http struct {
	Timeout    int
	ReportsUrl string
	TryCount   int
}

// Config define application configuration
type Config struct {
	Db   Db
	Http Http
}
