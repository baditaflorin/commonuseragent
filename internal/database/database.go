package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite" // SQLite driver
)

var (
	// ErrNotFound is returned when a record is not found
	ErrNotFound = errors.New("record not found")
	// ErrInvalidInput is returned when input validation fails
	ErrInvalidInput = errors.New("invalid input")
)

// DB wraps the database connection with additional functionality
type DB struct {
	conn *sql.DB
}

// RequestLog represents a user agent request log entry
type RequestLog struct {
	ID          int64
	UserAgent   string
	AgentType   string // "desktop", "mobile", or "random"
	RequestedAt time.Time
	IPAddress   string
	Endpoint    string
}

// Stats represents aggregated statistics
type Stats struct {
	TotalRequests   int64
	DesktopRequests int64
	MobileRequests  int64
	RandomRequests  int64
	UniqueIPs       int64
	LastRequest     time.Time
}

// New creates a new database connection
func New(dataSourceName string, maxOpenConns, maxIdleConns int, connMaxLifetime time.Duration) (*DB, error) {
	if dataSourceName == "" {
		return nil, fmt.Errorf("%w: data source name cannot be empty", ErrInvalidInput)
	}

	db, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	wrapper := &DB{conn: db}

	// Initialize schema
	if err := wrapper.initSchema(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return wrapper, nil
}

// initSchema creates the database schema if it doesn't exist
func (db *DB) initSchema(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS request_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_agent TEXT NOT NULL,
		agent_type TEXT NOT NULL,
		requested_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		ip_address TEXT NOT NULL,
		endpoint TEXT NOT NULL,
		CHECK(agent_type IN ('desktop', 'mobile', 'random')),
		CHECK(length(user_agent) <= 1000),
		CHECK(length(ip_address) <= 45),
		CHECK(length(endpoint) <= 255)
	);

	CREATE INDEX IF NOT EXISTS idx_requested_at ON request_logs(requested_at);
	CREATE INDEX IF NOT EXISTS idx_agent_type ON request_logs(agent_type);
	CREATE INDEX IF NOT EXISTS idx_ip_address ON request_logs(ip_address);
	`

	_, err := db.conn.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

// LogRequest logs a user agent request using parameterized queries
func (db *DB) LogRequest(ctx context.Context, log RequestLog) (int64, error) {
	// Validate input
	if err := validateRequestLog(&log); err != nil {
		return 0, err
	}

	// Use parameterized query to prevent SQL injection
	query := `
		INSERT INTO request_logs (user_agent, agent_type, requested_at, ip_address, endpoint)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := db.conn.ExecContext(ctx, query,
		log.UserAgent,
		log.AgentType,
		log.RequestedAt,
		log.IPAddress,
		log.Endpoint,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert request log: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

// GetRecentRequests retrieves the most recent N requests using parameterized queries
func (db *DB) GetRecentRequests(ctx context.Context, limit int) ([]RequestLog, error) {
	if limit < 1 || limit > 1000 {
		return nil, fmt.Errorf("%w: limit must be between 1 and 1000", ErrInvalidInput)
	}

	query := `
		SELECT id, user_agent, agent_type, requested_at, ip_address, endpoint
		FROM request_logs
		ORDER BY requested_at DESC
		LIMIT ?
	`

	rows, err := db.conn.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent requests: %w", err)
	}
	defer rows.Close()

	var logs []RequestLog
	for rows.Next() {
		var log RequestLog
		err := rows.Scan(
			&log.ID,
			&log.UserAgent,
			&log.AgentType,
			&log.RequestedAt,
			&log.IPAddress,
			&log.Endpoint,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return logs, nil
}

// GetRequestsByType retrieves requests by agent type using parameterized queries
func (db *DB) GetRequestsByType(ctx context.Context, agentType string, limit int) ([]RequestLog, error) {
	// Validate agent type
	validTypes := map[string]bool{"desktop": true, "mobile": true, "random": true}
	if !validTypes[agentType] {
		return nil, fmt.Errorf("%w: invalid agent type: %s", ErrInvalidInput, agentType)
	}

	if limit < 1 || limit > 1000 {
		return nil, fmt.Errorf("%w: limit must be between 1 and 1000", ErrInvalidInput)
	}

	query := `
		SELECT id, user_agent, agent_type, requested_at, ip_address, endpoint
		FROM request_logs
		WHERE agent_type = ?
		ORDER BY requested_at DESC
		LIMIT ?
	`

	rows, err := db.conn.QueryContext(ctx, query, agentType, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query requests by type: %w", err)
	}
	defer rows.Close()

	var logs []RequestLog
	for rows.Next() {
		var log RequestLog
		err := rows.Scan(
			&log.ID,
			&log.UserAgent,
			&log.AgentType,
			&log.RequestedAt,
			&log.IPAddress,
			&log.Endpoint,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return logs, nil
}

// GetStats retrieves aggregated statistics
func (db *DB) GetStats(ctx context.Context) (*Stats, error) {
	query := `
		SELECT
			COUNT(*) as total_requests,
			SUM(CASE WHEN agent_type = 'desktop' THEN 1 ELSE 0 END) as desktop_requests,
			SUM(CASE WHEN agent_type = 'mobile' THEN 1 ELSE 0 END) as mobile_requests,
			SUM(CASE WHEN agent_type = 'random' THEN 1 ELSE 0 END) as random_requests,
			COUNT(DISTINCT ip_address) as unique_ips,
			MAX(requested_at) as last_request
		FROM request_logs
	`

	var stats Stats
	var lastRequest sql.NullTime

	err := db.conn.QueryRowContext(ctx, query).Scan(
		&stats.TotalRequests,
		&stats.DesktopRequests,
		&stats.MobileRequests,
		&stats.RandomRequests,
		&stats.UniqueIPs,
		&lastRequest,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query stats: %w", err)
	}

	if lastRequest.Valid {
		stats.LastRequest = lastRequest.Time
	}

	return &stats, nil
}

// DeleteOldRequests deletes requests older than the specified duration using parameterized queries
func (db *DB) DeleteOldRequests(ctx context.Context, olderThan time.Duration) (int64, error) {
	if olderThan < 0 {
		return 0, fmt.Errorf("%w: duration cannot be negative", ErrInvalidInput)
	}

	cutoff := time.Now().Add(-olderThan)

	query := `DELETE FROM request_logs WHERE requested_at < ?`

	result, err := db.conn.ExecContext(ctx, query, cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old requests: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return count, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// Ping verifies the database connection is alive
func (db *DB) Ping(ctx context.Context) error {
	return db.conn.PingContext(ctx)
}

// validateRequestLog validates a RequestLog before insertion
func validateRequestLog(log *RequestLog) error {
	if log == nil {
		return fmt.Errorf("%w: log cannot be nil", ErrInvalidInput)
	}

	if log.UserAgent == "" {
		return fmt.Errorf("%w: user agent cannot be empty", ErrInvalidInput)
	}

	if len(log.UserAgent) > 1000 {
		return fmt.Errorf("%w: user agent exceeds maximum length of 1000 characters", ErrInvalidInput)
	}

	validTypes := map[string]bool{"desktop": true, "mobile": true, "random": true}
	if !validTypes[log.AgentType] {
		return fmt.Errorf("%w: invalid agent type: %s", ErrInvalidInput, log.AgentType)
	}

	if log.IPAddress == "" {
		return fmt.Errorf("%w: IP address cannot be empty", ErrInvalidInput)
	}

	if len(log.IPAddress) > 45 {
		return fmt.Errorf("%w: IP address exceeds maximum length", ErrInvalidInput)
	}

	if log.Endpoint == "" {
		return fmt.Errorf("%w: endpoint cannot be empty", ErrInvalidInput)
	}

	if len(log.Endpoint) > 255 {
		return fmt.Errorf("%w: endpoint exceeds maximum length", ErrInvalidInput)
	}

	if log.RequestedAt.IsZero() {
		log.RequestedAt = time.Now()
	}

	return nil
}
