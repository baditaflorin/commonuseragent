package commonuseragent

import (
	"errors"
	"strings"
	"sync"
	"testing"
)

func TestGetAllDesktop(t *testing.T) {
	desktops, err := GetAllDesktop()
	if err != nil {
		t.Fatalf("GetAllDesktop failed: %v", err)
	}
	if len(desktops) == 0 {
		t.Error("GetAllDesktop returned an empty slice")
	}

	// Verify we get a copy, not the original
	original := desktops[0].UA
	desktops[0].UA = "modified"
	desktops2, _ := GetAllDesktop()
	if desktops2[0].UA != original {
		t.Error("GetAllDesktop should return a copy, not the original slice")
	}
}

func TestGetAllMobile(t *testing.T) {
	mobiles, err := GetAllMobile()
	if err != nil {
		t.Fatalf("GetAllMobile failed: %v", err)
	}
	if len(mobiles) == 0 {
		t.Error("GetAllMobile returned an empty slice")
	}

	// Verify we get a copy, not the original
	original := mobiles[0].UA
	mobiles[0].UA = "modified"
	mobiles2, _ := GetAllMobile()
	if mobiles2[0].UA != original {
		t.Error("GetAllMobile should return a copy, not the original slice")
	}
}

func TestGetRandomDesktop(t *testing.T) {
	result, err := GetRandomDesktop()
	if err != nil {
		t.Fatalf("GetRandomDesktop failed: %v", err)
	}
	if result.UA == "" {
		t.Error("GetRandomDesktop returned an empty user agent")
	}
	if result.Pct <= 0 {
		t.Error("GetRandomDesktop returned invalid percentage")
	}
}

func TestGetRandomDesktopUA(t *testing.T) {
	result, err := GetRandomDesktopUA()
	if err != nil {
		t.Fatalf("GetRandomDesktopUA failed: %v", err)
	}
	if result == "" {
		t.Error("GetRandomDesktopUA returned an empty user agent")
	}
	if len(result) < 10 {
		t.Error("GetRandomDesktopUA returned suspiciously short user agent")
	}
}

func TestGetRandomMobile(t *testing.T) {
	result, err := GetRandomMobile()
	if err != nil {
		t.Fatalf("GetRandomMobile failed: %v", err)
	}
	if result.UA == "" {
		t.Error("GetRandomMobile returned an empty user agent")
	}
	if result.Pct <= 0 {
		t.Error("GetRandomMobile returned invalid percentage")
	}
}

func TestGetRandomMobileUA(t *testing.T) {
	result, err := GetRandomMobileUA()
	if err != nil {
		t.Fatalf("GetRandomMobileUA failed: %v", err)
	}
	if result == "" {
		t.Error("GetRandomMobileUA returned an empty user agent")
	}
	if len(result) < 10 {
		t.Error("GetRandomMobileUA returned suspiciously short user agent")
	}
}

func TestGetRandomUA(t *testing.T) {
	result, err := GetRandomUA()
	if err != nil {
		t.Fatalf("GetRandomUA failed: %v", err)
	}
	if result == "" {
		t.Error("GetRandomUA returned an empty user agent")
	}
}

// Security Tests

func TestCryptoRandomness(t *testing.T) {
	// Test that we get different values (not using a fixed seed)
	results := make(map[string]bool)
	iterations := 100

	for i := 0; i < iterations; i++ {
		ua, err := GetRandomUA()
		if err != nil {
			t.Fatalf("GetRandomUA failed: %v", err)
		}
		results[ua] = true
	}

	// We should see some variety (at least 2 different UAs in 100 tries)
	if len(results) < 2 {
		t.Error("Random function appears to be producing insufficient randomness")
	}
}

func TestThreadSafety(t *testing.T) {
	const goroutines = 100
	const iterations = 100

	var wg sync.WaitGroup
	errors := make(chan error, goroutines*iterations)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_, err := GetRandomUA()
				if err != nil {
					errors <- err
				}
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent access error: %v", err)
	}
}

func TestDataValidation(t *testing.T) {
	// Test that all loaded user agents are valid
	desktops, err := GetAllDesktop()
	if err != nil {
		t.Fatalf("Failed to get desktop agents: %v", err)
	}

	for i, agent := range desktops {
		if err := validateAgent(agent); err != nil {
			t.Errorf("Invalid desktop agent at index %d: %v", i, err)
		}
	}

	mobiles, err := GetAllMobile()
	if err != nil {
		t.Fatalf("Failed to get mobile agents: %v", err)
	}

	for i, agent := range mobiles {
		if err := validateAgent(agent); err != nil {
			t.Errorf("Invalid mobile agent at index %d: %v", i, err)
		}
	}
}

// Test Manager directly

func TestManagerCreation(t *testing.T) {
	mgr, err := NewManager(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	if mgr == nil {
		t.Fatal("Manager is nil")
	}
}

func TestManagerInvalidConfig(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "empty desktop file",
			config: Config{
				DesktopFile: "",
				MobileFile:  "mobile_useragents.json",
			},
		},
		{
			name: "empty mobile file",
			config: Config{
				DesktopFile: "desktop_useragents.json",
				MobileFile:  "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, err := NewManager(tt.config)
			if err == nil {
				t.Error("Expected error for invalid config, got nil")
			}
			if mgr != nil {
				t.Error("Expected nil manager for invalid config")
			}
		})
	}
}

func TestManagerNonExistentFile(t *testing.T) {
	cfg := Config{
		DesktopFile: "nonexistent_desktop.json",
		MobileFile:  "mobile_useragents.json",
	}

	mgr, err := NewManager(cfg)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
	if mgr != nil {
		t.Error("Expected nil manager for failed initialization")
	}
	if !errors.Is(err, ErrFileNotFound) {
		t.Errorf("Expected ErrFileNotFound, got: %v", err)
	}
}

func TestValidateAgent(t *testing.T) {
	tests := []struct {
		name      string
		agent     UserAgent
		shouldErr bool
	}{
		{
			name:      "valid agent",
			agent:     UserAgent{UA: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)", Pct: 50.0},
			shouldErr: false,
		},
		{
			name:      "empty UA",
			agent:     UserAgent{UA: "", Pct: 50.0},
			shouldErr: true,
		},
		{
			name:      "negative percentage",
			agent:     UserAgent{UA: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)", Pct: -1.0},
			shouldErr: true,
		},
		{
			name:      "percentage over 100",
			agent:     UserAgent{UA: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)", Pct: 101.0},
			shouldErr: true,
		},
		{
			name:      "UA too short",
			agent:     UserAgent{UA: "short", Pct: 50.0},
			shouldErr: true,
		},
		{
			name:      "UA too long",
			agent:     UserAgent{UA: strings.Repeat("a", 1001), Pct: 50.0},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAgent(tt.agent)
			if tt.shouldErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestSecureRandomInt(t *testing.T) {
	tests := []struct {
		name      string
		max       int
		shouldErr bool
	}{
		{
			name:      "valid max",
			max:       100,
			shouldErr: false,
		},
		{
			name:      "zero max",
			max:       0,
			shouldErr: true,
		},
		{
			name:      "negative max",
			max:       -1,
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := secureRandomInt(tt.max)
			if tt.shouldErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.shouldErr && (result < 0 || result >= tt.max) {
				t.Errorf("Result %d out of range [0, %d)", result, tt.max)
			}
		})
	}
}

func TestSecureRandomIntDistribution(t *testing.T) {
	// Test that secureRandomInt has reasonable distribution
	max := 10
	iterations := 1000
	counts := make([]int, max)

	for i := 0; i < iterations; i++ {
		n, err := secureRandomInt(max)
		if err != nil {
			t.Fatalf("secureRandomInt failed: %v", err)
		}
		counts[n]++
	}

	// Each bucket should have been hit at least once in 1000 iterations
	for i, count := range counts {
		if count == 0 {
			t.Errorf("Bucket %d was never selected in %d iterations", i, iterations)
		}
	}
}

// Benchmark tests

func BenchmarkGetRandomUA(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GetRandomUA()
		if err != nil {
			b.Fatalf("GetRandomUA failed: %v", err)
		}
	}
}

func BenchmarkGetRandomDesktop(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GetRandomDesktop()
		if err != nil {
			b.Fatalf("GetRandomDesktop failed: %v", err)
		}
	}
}

func BenchmarkGetRandomMobile(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GetRandomMobile()
		if err != nil {
			b.Fatalf("GetRandomMobile failed: %v", err)
		}
	}
}

func BenchmarkConcurrentAccess(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := GetRandomUA()
			if err != nil {
				b.Fatalf("GetRandomUA failed: %v", err)
			}
		}
	})
}
