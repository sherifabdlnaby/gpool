package StringReverser

import (
	"fmt"
	"log"
	"pipeline/Payload"
	"pipeline/Plugin"
)

var ProcessIdentifier = Plugin.ProcessIdentifier{
	Name:    "String Reverser",
	Version: "1.0.1",
}

type StringReverser struct {
}

func (s StringReverser) GetIdentifier() Plugin.ProcessIdentifier {
	return ProcessIdentifier
}

func (s StringReverser) Init(jsonConfig string) {
	log.Println("Setting Up Configs: ", jsonConfig)
}

func (s StringReverser) Start() {
	log.Println(fmt.Sprintf("STARTING Plugin: %s v%s", s.GetIdentifier().Name, s.GetIdentifier().Version))
}

func (s StringReverser) Run(p Payload.Payload) Payload.Payload {
	log.Println(fmt.Sprintf("RUNNING A PROCESS AT Plugin: %s v%s", s.GetIdentifier().Name, s.GetIdentifier().Version))
	return Payload.Payload{
		Json:  p.Json,
		Img:   p.Img,
		Count: p.Count + 1,
	}
}

func (s StringReverser) Close() {
	log.Println(fmt.Sprintf("CLOSING Plugin: %s v%s", s.GetIdentifier().Name, s.GetIdentifier().Version))

}
