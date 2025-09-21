-- Database initialization script for Prototype Game Backend
-- This script sets up the basic database schema for persistence

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_login TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT true
);

-- Create players table
CREATE TABLE IF NOT EXISTS players (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    player_name VARCHAR(50) NOT NULL,
    level INTEGER DEFAULT 1,
    experience BIGINT DEFAULT 0,
    position_x DOUBLE PRECISION DEFAULT 0,
    position_z DOUBLE PRECISION DEFAULT 0,
    health DOUBLE PRECISION DEFAULT 100,
    max_health DOUBLE PRECISION DEFAULT 100,
    skills JSONB DEFAULT '{}',
    inventory JSONB DEFAULT '{}',
    equipment JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_active TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create game sessions table
CREATE TABLE IF NOT EXISTS game_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    session_token VARCHAR(255) UNIQUE NOT NULL,
    ip_address INET,
    user_agent TEXT,
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ended_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT true
);

-- Create game events table (for analytics and debugging)
CREATE TABLE IF NOT EXISTS game_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    player_id UUID REFERENCES players(id) ON DELETE SET NULL,
    session_id UUID REFERENCES game_sessions(id) ON DELETE SET NULL,
    event_type VARCHAR(50) NOT NULL,
    event_data JSONB NOT NULL,
    position_x DOUBLE PRECISION,
    position_z DOUBLE PRECISION,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create spatial cells table (for world partitioning)
CREATE TABLE IF NOT EXISTS spatial_cells (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cell_x INTEGER NOT NULL,
    cell_z INTEGER NOT NULL,
    cell_size DOUBLE PRECISION NOT NULL,
    entity_count INTEGER DEFAULT 0,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(cell_x, cell_z, cell_size)
);

-- Create item templates table
CREATE TABLE IF NOT EXISTS item_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    template_id VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    item_type VARCHAR(50) NOT NULL,
    rarity VARCHAR(20) DEFAULT 'common',
    base_stats JSONB DEFAULT '{}',
    requirements JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create world state table (for persistent world data)
CREATE TABLE IF NOT EXISTS world_state (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    region_id VARCHAR(100) NOT NULL,
    state_data JSONB NOT NULL,
    version INTEGER DEFAULT 1,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(region_id)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_players_user_id ON players(user_id);
CREATE INDEX IF NOT EXISTS idx_players_position ON players(position_x, position_z);
CREATE INDEX IF NOT EXISTS idx_players_last_active ON players(last_active);

CREATE INDEX IF NOT EXISTS idx_game_sessions_player_id ON game_sessions(player_id);
CREATE INDEX IF NOT EXISTS idx_game_sessions_active ON game_sessions(is_active);
CREATE INDEX IF NOT EXISTS idx_game_sessions_token ON game_sessions(session_token);

CREATE INDEX IF NOT EXISTS idx_game_events_player_id ON game_events(player_id);
CREATE INDEX IF NOT EXISTS idx_game_events_type ON game_events(event_type);
CREATE INDEX IF NOT EXISTS idx_game_events_timestamp ON game_events(timestamp);
CREATE INDEX IF NOT EXISTS idx_game_events_position ON game_events(position_x, position_z);

CREATE INDEX IF NOT EXISTS idx_spatial_cells_position ON spatial_cells(cell_x, cell_z);
CREATE INDEX IF NOT EXISTS idx_spatial_cells_updated ON spatial_cells(last_updated);

-- Functions for automatic timestamp updates
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers for automatic timestamp updates
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_players_updated_at ON players;
CREATE TRIGGER update_players_updated_at BEFORE UPDATE ON players 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_item_templates_updated_at ON item_templates;
CREATE TRIGGER update_item_templates_updated_at BEFORE UPDATE ON item_templates 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_world_state_updated_at ON world_state;
CREATE TRIGGER update_world_state_updated_at BEFORE UPDATE ON world_state 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert some sample data for development
INSERT INTO item_templates (template_id, name, description, item_type, rarity, base_stats, requirements)
VALUES 
    ('sword_iron', 'Iron Sword', 'A sturdy iron sword for combat', 'weapon', 'common', 
     '{"damage": 25, "durability": 100}', '{"melee": 5}'),
    ('armor_leather', 'Leather Armor', 'Basic leather protection', 'armor', 'common',
     '{"defense": 15, "durability": 80}', '{"defense": 3}'),
    ('potion_health', 'Health Potion', 'Restores health when consumed', 'consumable', 'common',
     '{"heal_amount": 50}', '{}')
ON CONFLICT (template_id) DO NOTHING;

-- Create a view for active players with session info
CREATE OR REPLACE VIEW active_players AS
SELECT 
    p.*,
    s.session_token,
    s.started_at as session_started,
    s.ip_address
FROM players p
JOIN game_sessions s ON p.id = s.player_id
WHERE s.is_active = true
  AND s.ended_at IS NULL;

-- Create a view for player statistics
CREATE OR REPLACE VIEW player_statistics AS
SELECT 
    p.id,
    p.player_name,
    p.level,
    p.experience,
    COUNT(e.id) as total_events,
    MAX(e.timestamp) as last_event,
    EXTRACT(EPOCH FROM (NOW() - p.created_at)) / 3600 as hours_played
FROM players p
LEFT JOIN game_events e ON p.id = e.player_id
GROUP BY p.id, p.player_name, p.level, p.experience, p.created_at;

-- Grants for application user (if needed)
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO gameuser;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO gameuser;

COMMIT;