package commonuseragent

import (
	"crypto/rand"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sync"
)

// Go directive to embed the files in the binary.
//
//go:embed desktop_useragents.json
//go:embed mobile_useragents.json
var content embed.FS

// UserAgent represents a user agent string with its usage percentage
type UserAgent struct {
	UA  string  `json:"ua"`
	Pct float64 `json:"pct"`
}

// Config holds runtime configuration for the user agent library
type Config struct {
	DesktopFile string
	MobileFile  string
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		DesktopFile: "desktop_useragents.json",
		MobileFile:  "mobile_useragents.json",
	}
}

// Manager handles user agent data with thread-safe operations
type Manager struct {
	mu            sync.RWMutex
	desktopAgents []UserAgent
	mobileAgents  []UserAgent
	config        Config
}

var (
	// ErrEmptyAgentList is returned when trying to get a random agent from an empty list
	ErrEmptyAgentList = errors.New("agent list is empty")
	// ErrInvalidData is returned when user agent data is invalid
	ErrInvalidData = errors.New("invalid user agent data")
	// ErrFileNotFound is returned when the embedded file cannot be found
	ErrFileNotFound = errors.New("embedded file not found")
)

var (
	defaultManager     *Manager
	defaultManagerOnce sync.Once
	initError          error
)

// init initializes the default manager with error handling instead of panic
func init() {
	defaultManagerOnce.Do(func() {
		mgr, err := NewManager(DefaultConfig())
		if err != nil {
			// Store error for later retrieval instead of panic
			initError = err
			return
		}
		defaultManager = mgr
	})
}

// GetInitError returns any error that occurred during initialization
func GetInitError() error {
	return initError
}

// NewManager creates a new Manager with the given configuration
func NewManager(cfg Config) (*Manager, error) {
	if cfg.DesktopFile == "" || cfg.MobileFile == "" {
		return nil, fmt.Errorf("%w: desktop and mobile files must be specified", ErrInvalidData)
	}

	m := &Manager{
		config: cfg,
	}

	if err := m.loadUserAgents(cfg.DesktopFile, &m.desktopAgents); err != nil {
		return nil, fmt.Errorf("failed to load desktop agents: %w", err)
	}

	if err := m.loadUserAgents(cfg.MobileFile, &m.mobileAgents); err != nil {
		return nil, fmt.Errorf("failed to load mobile agents: %w", err)
	}

	// Validate loaded data
	if err := m.validate(); err != nil {
		return nil, err
	}

	return m, nil
}

// loadUserAgents reads and unmarshals user agent data from embedded files
func (m *Manager) loadUserAgents(filename string, agents *[]UserAgent) error {
	bytes, err := content.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrFileNotFound, filename)
	}

	if err := json.Unmarshal(bytes, agents); err != nil {
		return fmt.Errorf("failed to parse %s: %w", filename, err)
	}

	return nil
}

// validate ensures the loaded user agent data is valid
func (m *Manager) validate() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.desktopAgents) == 0 && len(m.mobileAgents) == 0 {
		return fmt.Errorf("%w: both desktop and mobile agent lists are empty", ErrInvalidData)
	}

	// Validate individual agents
	for i, agent := range m.desktopAgents {
		if err := validateAgent(agent); err != nil {
			return fmt.Errorf("invalid desktop agent at index %d: %w", i, err)
		}
	}

	for i, agent := range m.mobileAgents {
		if err := validateAgent(agent); err != nil {
			return fmt.Errorf("invalid mobile agent at index %d: %w", i, err)
		}
	}

	return nil
}

// validateAgent checks if a single UserAgent is valid
func validateAgent(ua UserAgent) error {
	if ua.UA == "" {
		return fmt.Errorf("%w: user agent string is empty", ErrInvalidData)
	}
	if ua.Pct < 0 || ua.Pct > 100 {
		return fmt.Errorf("%w: percentage must be between 0 and 100, got %.2f", ErrInvalidData, ua.Pct)
	}
	// Basic sanity check for UA string length
	if len(ua.UA) < 10 || len(ua.UA) > 1000 {
		return fmt.Errorf("%w: user agent string length must be between 10 and 1000 characters", ErrInvalidData)
	}
	return nil
}

// GetAllDesktop returns a copy of all desktop user agents
func (m *Manager) GetAllDesktop() []UserAgent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent external modification
	agents := make([]UserAgent, len(m.desktopAgents))
	copy(agents, m.desktopAgents)
	return agents
}

// GetAllMobile returns a copy of all mobile user agents
func (m *Manager) GetAllMobile() []UserAgent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent external modification
	agents := make([]UserAgent, len(m.mobileAgents))
	copy(agents, m.mobileAgents)
	return agents
}

// GetRandomDesktop returns a random desktop UserAgent using crypto/rand
func (m *Manager) GetRandomDesktop() (UserAgent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.desktopAgents) == 0 {
		return UserAgent{}, ErrEmptyAgentList
	}

	idx, err := secureRandomInt(len(m.desktopAgents))
	if err != nil {
		return UserAgent{}, fmt.Errorf("failed to generate random index: %w", err)
	}

	return m.desktopAgents[idx], nil
}

// GetRandomMobile returns a random mobile UserAgent using crypto/rand
func (m *Manager) GetRandomMobile() (UserAgent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.mobileAgents) == 0 {
		return UserAgent{}, ErrEmptyAgentList
	}

	idx, err := secureRandomInt(len(m.mobileAgents))
	if err != nil {
		return UserAgent{}, fmt.Errorf("failed to generate random index: %w", err)
	}

	return m.mobileAgents[idx], nil
}

// GetRandomDesktopUA returns just the UA string of a random desktop user agent
func (m *Manager) GetRandomDesktopUA() (string, error) {
	ua, err := m.GetRandomDesktop()
	if err != nil {
		return "", err
	}
	return ua.UA, nil
}

// GetRandomMobileUA returns just the UA string of a random mobile user agent
func (m *Manager) GetRandomMobileUA() (string, error) {
	ua, err := m.GetRandomMobile()
	if err != nil {
		return "", err
	}
	return ua.UA, nil
}

// GetRandomUA returns a random user agent string from either desktop or mobile
func (m *Manager) GetRandomUA() (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	allAgents := make([]UserAgent, 0, len(m.desktopAgents)+len(m.mobileAgents))
	allAgents = append(allAgents, m.desktopAgents...)
	allAgents = append(allAgents, m.mobileAgents...)

	if len(allAgents) == 0 {
		return "", ErrEmptyAgentList
	}

	idx, err := secureRandomInt(len(allAgents))
	if err != nil {
		return "", fmt.Errorf("failed to generate random index: %w", err)
	}

	return allAgents[idx].UA, nil
}

// secureRandomInt generates a cryptographically secure random integer in [0, max)
func secureRandomInt(max int) (int, error) {
	if max <= 0 {
		return 0, errors.New("max must be positive")
	}

	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}

	return int(nBig.Int64()), nil
}

// Package-level convenience functions that use the default manager

// GetAllDesktop returns all desktop user agents using the default manager
func GetAllDesktop() ([]UserAgent, error) {
	if initError != nil {
		return nil, fmt.Errorf("library not initialized: %w", initError)
	}
	return defaultManager.GetAllDesktop(), nil
}

// GetAllMobile returns all mobile user agents using the default manager
func GetAllMobile() ([]UserAgent, error) {
	if initError != nil {
		return nil, fmt.Errorf("library not initialized: %w", initError)
	}
	return defaultManager.GetAllMobile(), nil
}

// GetRandomDesktop returns a random desktop UserAgent using the default manager
func GetRandomDesktop() (UserAgent, error) {
	if initError != nil {
		return UserAgent{}, fmt.Errorf("library not initialized: %w", initError)
	}
	return defaultManager.GetRandomDesktop()
}

// GetRandomMobile returns a random mobile UserAgent using the default manager
func GetRandomMobile() (UserAgent, error) {
	if initError != nil {
		return UserAgent{}, fmt.Errorf("library not initialized: %w", initError)
	}
	return defaultManager.GetRandomMobile()
}

// GetRandomDesktopUA returns just the UA string of a random desktop user agent using the default manager
func GetRandomDesktopUA() (string, error) {
	if initError != nil {
		return "", fmt.Errorf("library not initialized: %w", initError)
	}
	return defaultManager.GetRandomDesktopUA()
}

// GetRandomMobileUA returns just the UA string of a random mobile user agent using the default manager
func GetRandomMobileUA() (string, error) {
	if initError != nil {
		return "", fmt.Errorf("library not initialized: %w", initError)
	}
	return defaultManager.GetRandomMobileUA()
}

// GetRandomUA returns a random user agent string from either desktop or mobile using the default manager
func GetRandomUA() (string, error) {
	if initError != nil {
		return "", fmt.Errorf("library not initialized: %w", initError)
	}
	return defaultManager.GetRandomUA()
}
