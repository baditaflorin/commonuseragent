package commonuseragent

import (
	"math"
	"testing"
)

func TestWeightedRandomSelection(t *testing.T) {
	// Use a custom manager with known weights for testing
	cfg := Config{
		DesktopFile: "desktop_useragents.json", // We'll override the data anyway
		MobileFile:  "mobile_useragents.json",
	}
	mgr, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Override agents with known weights
	mgr.desktopAgents = []UserAgent{
		{UA: "A", Pct: 10.0},
		{UA: "B", Pct: 30.0},
		{UA: "C", Pct: 60.0},
	}

	iterations := 10000
	counts := make(map[string]int)

	for i := 0; i < iterations; i++ {
		ua, err := mgr.GetRandomDesktop()
		if err != nil {
			t.Fatalf("GetRandomDesktop failed: %v", err)
		}
		counts[ua.UA]++
	}

	// Check if distribution is roughly correct (within 5% margin of error)
	expected := map[string]float64{
		"A": 0.10,
		"B": 0.30,
		"C": 0.60,
	}

	for ua, count := range counts {
		observedPct := float64(count) / float64(iterations)
		expectedPct := expected[ua]
		diff := math.Abs(observedPct - expectedPct)

		// Allow 2% deviation
		if diff > 0.02 {
			t.Errorf("Distribution for %s incorrect. Expected %.2f, got %.2f", ua, expectedPct, observedPct)
		}
	}
}
