package main

import "time"

const ServerID = 0

var Version string

var ServerInfo struct {
	Fingerprint string
	StartTime   time.Time
}
