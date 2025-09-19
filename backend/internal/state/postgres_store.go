package state

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

var (
	ErrOptimisticLock = errors.New("optimistic locking conflict - state was modified by another process")
	ErrPlayerNotFound = errors.New("player not found")
)

// PostgresStore implements persistence using PostgreSQL with optimistic locking
type PostgresStore struct {
	db       *sql.DB
	saveStmt *sql.Stmt
	loadStmt *sql.Stmt
}

// NewPostgresStore creates a new PostgreSQL-backed store
func NewPostgresStore(dsn string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool for optimal performance
	db.SetMaxOpenConns(25)                  // Maximum number of open connections
	db.SetMaxIdleConns(10)                  // Maximum number of idle connections
	db.SetConnMaxLifetime(30 * time.Minute) // Maximum connection lifetime

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	store := &PostgresStore{db: db}

	if err := store.createSchema(); err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	if err := store.prepareStatements(); err != nil {
		return nil, fmt.Errorf("failed to prepare statements: %w", err)
	}

	return store, nil
}

// createSchema creates the required tables if they don't exist
func (ps *PostgresStore) createSchema() error {
	schema := `
		CREATE TABLE IF NOT EXISTS player_state (
			player_id TEXT PRIMARY KEY,
			pos_x FLOAT8 NOT NULL DEFAULT 0,
			pos_z FLOAT8 NOT NULL DEFAULT 0,
			logins INTEGER NOT NULL DEFAULT 0,
			updated TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			version BIGINT NOT NULL DEFAULT 1,
			inventory_data JSONB,
			equipment_data JSONB,
			skills_data JSONB,
			cooldown_timers JSONB,
			encumbrance_config JSONB,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		);

		-- Index for efficient lookups
		CREATE INDEX IF NOT EXISTS idx_player_state_updated ON player_state(updated);
		CREATE INDEX IF NOT EXISTS idx_player_state_version ON player_state(version);
	`

	_, err := ps.db.Exec(schema)
	return err
}

// prepareStatements prepares frequently used SQL statements
func (ps *PostgresStore) prepareStatements() error {
	var err error

	// Save statement with optimistic locking
	ps.saveStmt, err = ps.db.Prepare(`
		INSERT INTO player_state (
			player_id, pos_x, pos_z, logins, updated, version,
			inventory_data, equipment_data, skills_data, cooldown_timers, encumbrance_config
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (player_id) DO UPDATE SET
			pos_x = EXCLUDED.pos_x,
			pos_z = EXCLUDED.pos_z,
			logins = EXCLUDED.logins,
			updated = EXCLUDED.updated,
			version = player_state.version + 1,
			inventory_data = EXCLUDED.inventory_data,
			equipment_data = EXCLUDED.equipment_data,
			skills_data = EXCLUDED.skills_data,
			cooldown_timers = EXCLUDED.cooldown_timers,
			encumbrance_config = EXCLUDED.encumbrance_config
		WHERE player_state.version = $6
		RETURNING version
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare save statement: %w", err)
	}

	// Load statement
	ps.loadStmt, err = ps.db.Prepare(`
		SELECT pos_x, pos_z, logins, updated, version,
			   inventory_data, equipment_data, skills_data, cooldown_timers, encumbrance_config
		FROM player_state
		WHERE player_id = $1
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare load statement: %w", err)
	}

	return nil
}

// Load retrieves player state from PostgreSQL
func (ps *PostgresStore) Load(ctx context.Context, playerID string) (PlayerState, bool, error) {
	var state PlayerState
	var inventoryData, equipmentData, skillsData, cooldownTimers, encumbranceConfig sql.NullString

	err := ps.loadStmt.QueryRowContext(ctx, playerID).Scan(
		&state.Pos.X, &state.Pos.Z, &state.Logins, &state.Updated, &state.Version,
		&inventoryData, &equipmentData, &skillsData, &cooldownTimers, &encumbranceConfig,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return PlayerState{}, false, nil
		}
		return PlayerState{}, false, fmt.Errorf("failed to load player state: %w", err)
	}

	// Convert nullable strings to RawMessage
	if inventoryData.Valid {
		state.InventoryData = json.RawMessage(inventoryData.String)
	}
	if equipmentData.Valid {
		state.EquipmentData = json.RawMessage(equipmentData.String)
	}
	if skillsData.Valid {
		state.SkillsData = json.RawMessage(skillsData.String)
	}
	if cooldownTimers.Valid {
		state.CooldownTimers = json.RawMessage(cooldownTimers.String)
	}
	if encumbranceConfig.Valid {
		state.EncumbranceConfig = json.RawMessage(encumbranceConfig.String)
	}

	return state, true, nil
}

// Save persists player state to PostgreSQL with optimistic locking
func (ps *PostgresStore) Save(ctx context.Context, playerID string, st PlayerState) error {
	// Convert json.RawMessage to nullable strings for database storage
	var inventoryData, equipmentData, skillsData, cooldownTimers, encumbranceConfig sql.NullString

	if len(st.InventoryData) > 0 {
		inventoryData = sql.NullString{String: string(st.InventoryData), Valid: true}
	}
	if len(st.EquipmentData) > 0 {
		equipmentData = sql.NullString{String: string(st.EquipmentData), Valid: true}
	}
	if len(st.SkillsData) > 0 {
		skillsData = sql.NullString{String: string(st.SkillsData), Valid: true}
	}
	if len(st.CooldownTimers) > 0 {
		cooldownTimers = sql.NullString{String: string(st.CooldownTimers), Valid: true}
	}
	if len(st.EncumbranceConfig) > 0 {
		encumbranceConfig = sql.NullString{String: string(st.EncumbranceConfig), Valid: true}
	}

	var newVersion int64
	err := ps.saveStmt.QueryRowContext(ctx,
		playerID, st.Pos.X, st.Pos.Z, st.Logins, st.Updated, st.Version,
		inventoryData, equipmentData, skillsData, cooldownTimers, encumbranceConfig,
	).Scan(&newVersion)

	if err != nil {
		if err == sql.ErrNoRows {
			return ErrOptimisticLock
		}
		return fmt.Errorf("failed to save player state: %w", err)
	}

	return nil
}

// Close closes the database connection and prepared statements
func (ps *PostgresStore) Close() error {
	if ps.saveStmt != nil {
		ps.saveStmt.Close()
	}
	if ps.loadStmt != nil {
		ps.loadStmt.Close()
	}
	if ps.db != nil {
		return ps.db.Close()
	}
	return nil
}

// GetStats returns basic statistics about the store
func (ps *PostgresStore) GetStats(ctx context.Context) (map[string]interface{}, error) {
	var totalPlayers, totalLogins int64
	var avgLogins float64

	err := ps.db.QueryRowContext(ctx, `
		SELECT 
			COUNT(*) as total_players,
			COALESCE(SUM(logins), 0) as total_logins,
			COALESCE(AVG(logins), 0) as avg_logins
		FROM player_state
	`).Scan(&totalPlayers, &totalLogins, &avgLogins)

	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	return map[string]interface{}{
		"total_players": totalPlayers,
		"total_logins":  totalLogins,
		"avg_logins":    avgLogins,
	}, nil
}

// withTimeout creates a context with a reasonable timeout for database operations
func (ps *PostgresStore) withTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, 5*time.Second)
}

// LoadWithTimeout retrieves player state with a timeout
func (ps *PostgresStore) LoadWithTimeout(ctx context.Context, playerID string) (PlayerState, bool, error) {
	timeoutCtx, cancel := ps.withTimeout(ctx)
	defer cancel()
	return ps.Load(timeoutCtx, playerID)
}

// SaveWithTimeout persists player state with a timeout
func (ps *PostgresStore) SaveWithTimeout(ctx context.Context, playerID string, st PlayerState) error {
	timeoutCtx, cancel := ps.withTimeout(ctx)
	defer cancel()
	return ps.Save(timeoutCtx, playerID, st)
}
