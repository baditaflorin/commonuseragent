package commonuseragent

import (
	"embed"
	"encoding/json"
	"math/rand"
	"time"
)

// Go directive to embed the files in the binary.
//
//go:embed desktop_useragents.json
//go:embed mobile_useragents.json
var content embed.FS

type UserAgent struct {
	UA  string  `json:"ua"`
	Pct float64 `json:"pct"`
}

var desktopAgents []UserAgent
var mobileAgents []UserAgent

func init() {
	rand.Seed(time.Now().UnixNano())
	loadUserAgents("desktop_useragents.json", &desktopAgents)
	loadUserAgents("mobile_useragents.json", &mobileAgents)
}

func loadUserAgents(filename string, agents *[]UserAgent) {
	// Reading from the embedded file system
	bytes, err := content.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(bytes, agents); err != nil {
		panic(err)
	}
}

func GetAllDesktop() []UserAgent {
	return desktopAgents
}

func GetAllMobile() []UserAgent {
	return mobileAgents
}

// GetRandomDesktop returns a random UserAgent struct from the desktopAgents slice
func GetRandomDesktop() UserAgent {
	if len(desktopAgents) == 0 {
		return UserAgent{}
	}
	return desktopAgents[rand.Intn(len(desktopAgents))]
}

// GetRandomMobile returns a random UserAgent struct from the mobileAgents slice
func GetRandomMobile() UserAgent {
	if len(mobileAgents) == 0 {
		return UserAgent{}
	}
	return mobileAgents[rand.Intn(len(mobileAgents))]
}

// GetRandomDesktopUA returns just the UA string of a random desktop user agent
func GetRandomDesktopUA() string {
	return GetRandomDesktop().UA
}

// GetRandomMobileUA returns just the UA string of a random mobile user agent
func GetRandomMobileUA() string {
	return GetRandomMobile().UA
}
func GetRandomUA() string {
	allAgents := append(desktopAgents, mobileAgents...)
	return allAgents[rand.Intn(len(allAgents))].UA
}
