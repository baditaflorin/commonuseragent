package commonuseragent

import (
	"testing"
)

func TestGetAllDesktop(t *testing.T) {
	desktops := GetAllDesktop()
	if len(desktops) == 0 {
		t.Errorf("GetAllDesktop returned an empty slice")
	}
}

func TestGetAllMobile(t *testing.T) {
	mobiles := GetAllMobile()
	if len(mobiles) == 0 {
		t.Errorf("GetAllMobile returned an empty slice")
	}
}

func TestGetRandomDesktop(t *testing.T) {
	// Calling the function to test
	result := GetRandomDesktop()
	if result.UA == "" {
		t.Errorf("GetRandomDesktop returned an empty user agent")
	}
}

func TestGetRandomMobile(t *testing.T) {
	// Calling the function to test
	result := GetRandomMobile()
	if result.UA == "" {
		t.Errorf("GetRandomMobile returned an empty user agent")
	}
}

func TestGetRandomUserAgent(t *testing.T) {
	result := GetRandomUserAgent()
	if result.UA == "" {
		t.Errorf("GetRandomUserAgent returned an empty user agent")
	}
}
