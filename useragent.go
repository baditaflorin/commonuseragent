package commonuseragent

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"time"
)

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
	bytes, err := ioutil.ReadFile(filename)
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

func GetRandomDesktop() UserAgent {
	return desktopAgents[rand.Intn(len(desktopAgents))]
}

func GetRandomMobile() UserAgent {
	return mobileAgents[rand.Intn(len(mobileAgents))]
}

func GetRandomUserAgent() UserAgent {
	allAgents := append(desktopAgents, mobileAgents...)
	return allAgents[rand.Intn(len(allAgents))]
}
