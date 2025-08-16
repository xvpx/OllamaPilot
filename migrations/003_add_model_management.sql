-- Migration: Add model management tables
-- Created: 2025-08-16

-- Models table for storing model metadata
CREATE TABLE IF NOT EXISTS models (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    description TEXT DEFAULT '',
    size INTEGER DEFAULT 0,
    family TEXT DEFAULT '',
    format TEXT DEFAULT '',
    parameters TEXT DEFAULT '',
    quantization TEXT DEFAULT '',
    status TEXT NOT NULL DEFAULT 'available' CHECK (status IN ('available', 'downloading', 'installing', 'error', 'removed')),
    is_default BOOLEAN DEFAULT FALSE,
    is_enabled BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_used_at DATETIME
);

-- Model configurations table
CREATE TABLE IF NOT EXISTS model_configs (
    id TEXT PRIMARY KEY,
    model_id TEXT NOT NULL,
    temperature REAL,
    top_p REAL,
    top_k INTEGER,
    repeat_penalty REAL,
    context_length INTEGER,
    max_tokens INTEGER,
    system_prompt TEXT DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE CASCADE
);

-- Model usage statistics view
CREATE VIEW IF NOT EXISTS model_usage_stats AS
SELECT 
    m.id as model_id,
    m.name,
    COUNT(msg.id) as total_messages,
    COALESCE(SUM(msg.tokens_used), 0) as total_tokens,
    MAX(msg.created_at) as last_used_at
FROM models m
LEFT JOIN messages msg ON m.name = msg.model
GROUP BY m.id, m.name;

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_models_name ON models(name);
CREATE INDEX IF NOT EXISTS idx_models_status ON models(status);
CREATE INDEX IF NOT EXISTS idx_models_is_enabled ON models(is_enabled);
CREATE INDEX IF NOT EXISTS idx_models_is_default ON models(is_default);
CREATE INDEX IF NOT EXISTS idx_models_last_used ON models(last_used_at);
CREATE INDEX IF NOT EXISTS idx_model_configs_model_id ON model_configs(model_id);

-- Trigger to update updated_at timestamp
CREATE TRIGGER IF NOT EXISTS update_models_updated_at
    AFTER UPDATE ON models
    FOR EACH ROW
BEGIN
    UPDATE models SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_model_configs_updated_at
    AFTER UPDATE ON model_configs
    FOR EACH ROW
BEGIN
    UPDATE model_configs SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Trigger to ensure only one default model
CREATE TRIGGER IF NOT EXISTS ensure_single_default_model
    AFTER UPDATE OF is_default ON models
    FOR EACH ROW
    WHEN NEW.is_default = TRUE
BEGIN
    UPDATE models SET is_default = FALSE WHERE id != NEW.id AND is_default = TRUE;
END;

-- Trigger to update model last_used_at when used in messages
CREATE TRIGGER IF NOT EXISTS update_model_last_used
    AFTER INSERT ON messages
    FOR EACH ROW
    WHEN NEW.model IS NOT NULL
BEGIN
    UPDATE models 
    SET last_used_at = NEW.created_at 
    WHERE name = NEW.model;
END;