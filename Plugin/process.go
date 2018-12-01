package Plugin

import (
	"pipeline/Payload"
)

type ProcessIdentifier struct {
	Name    string
	Version string
}

type Process interface {
	GetIdentifier() ProcessIdentifier
	Init(jsonConfig string)
	Start()
	Run(p Payload.Payload) Payload.Payload
	Close()
}
