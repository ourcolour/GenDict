package services

import (
	"embed"
)

//go:embed templates/*
var templateFiles embed.FS
