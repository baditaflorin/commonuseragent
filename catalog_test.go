package commonuseragent

import (
	"strings"
	"testing"
)

func TestCatalogContainsCurrentBrowserFamilies(t *testing.T) {
	desktop, err := GetAllDesktop()
	if err != nil {
		t.Fatalf("GetAllDesktop failed: %v", err)
	}
	mobile, err := GetAllMobile()
	if err != nil {
		t.Fatalf("GetAllMobile failed: %v", err)
	}

	assertCatalogContains(t, desktop, "Chrome/151.0.0.0", "desktop Chrome")
	assertCatalogContains(t, desktop, "Edg/151.0.0.0", "desktop Edge")
	assertCatalogContains(t, desktop, "Firefox/152.0", "desktop Firefox")
	assertCatalogContains(t, desktop, "Version/26.6 Safari", "desktop Safari")
	assertCatalogContains(t, mobile, "Chrome/151.0.0.0 Mobile", "Android Chrome")
	assertCatalogContains(t, mobile, "Version/26.6 Mobile", "iPhone Safari")
	assertCatalogContains(t, mobile, "SamsungBrowser/30.0", "Samsung Internet")
}

func assertCatalogContains(t *testing.T, agents []UserAgent, want, label string) {
	t.Helper()
	for _, agent := range agents {
		if strings.Contains(agent.UA, want) {
			return
		}
	}
	t.Errorf("catalog is missing %s (%q)", label, want)
}
