package main

import "github.com/AminN77/senju/backend/internal/httpserver"

// Link-time metadata (-ldflags "-X main.version=...").
var (
	version   = "0.0.0-dev"
	commit    = "unknown"
	buildTime = "unknown"
)

func versionInfo() httpserver.VersionInfo {
	return httpserver.VersionInfo{
		Service:   "senju-api",
		Version:   version,
		Commit:    commit,
		BuildTime: buildTime,
	}
}
