package StringReverser

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log"
	"pipeline/Payload"
	"pipeline/Plugin"
	"time"
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
	log.Println(fmt.Sprintf("RUNNING A PROCESS AT Plugin: [%s] [v%s] with PAYLOAD: %s", s.GetIdentifier().Name, s.GetIdentifier().Version, p))
	time.Sleep(1500 * time.Millisecond)

	//DO SOME STUFF
	hasher := sha1.New()
	hasher.Write([]byte(p.Json))
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	//
	return Payload.Payload{
		Json:  sha,
		Img:   p.Img,
		Count: p.Count + 1,
	}
}

func (s StringReverser) Close() {
	log.Println(fmt.Sprintf("CLOSING Plugin: %s v%s", s.GetIdentifier().Name, s.GetIdentifier().Version))
}
