package main

import (
	"github.com/ipfs/go-log/v2"
)

func initLogger() {
	log.SetAllLoggers(log.LevelWarn)
	log.SetLogLevel("rendezvous", "info")
}
