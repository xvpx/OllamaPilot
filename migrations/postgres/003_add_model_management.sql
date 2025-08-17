-- Models table for storing model metadata
CREATE TABLE models (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    description TEXT DEFAULT '',
    size BIGINT DEFAULT 0,
    family TEXT DEFAULT '',
    format TEXT DEFAULT '',
    parameters TEXT DEFAULT '',
    quantization TEXT DEFAULT '',
    status TEXT NOT NULL DEFAULT 'available' CHECK (status IN ('available', 'downloading', 'installing', 'error', 'removed')),
    is_default BOOLEAN DEFAULT FALSE,
    is_enabled BOOLEAN DEFAULT TRUE,
    supports_embeddings BOOLEAN DEFAULT FALSE,
    embedding_dimensions INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP WITH TIME ZONE
);

-- Model configurations table
CREATE TABLE model_configs (
    id TEXT PRIMARY KEY,
    model_id TEXT NOT NULL,
    temperature REAL,
    top_p REAL,
    top_k INTEGER,
    repeat_penalty REAL,
    context_length INTEGER,
    max_tokens INTEGER,
    system_prompt TEXT DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE CASCADE
);

-- Model usage statistics view
CREATE VIEW model_usage_stats AS
SELECT 
    m.id as model_id,
    m.name,
    COUNT(msg.id) as total_messages,
    COALESCE(SUM(msg.tokens_used), 0) as total_tokens,
    MAX(msg.created_at) as last_used_at,
    COUNT(DISTINCT msg.session_id) as unique_sessions
FROM models m
LEFT JOIN messages msg ON m.name = msg.model
GROUP BY m.id, m.name;

-- Embedding model usage view
CREATE VIEW embedding_usage_stats AS
SELECT 
    me.model_used,
    COUNT(*) as total_embeddings,
    MAX(me.created_at) as last_used_at,
    COUNT(DISTINCT m.session_id) as unique_sessions
FROM message_embeddings me
JOIN messages m ON me.message_id = m.id
GROUP BY me.model_used;

-- Indexes for performance
CREATE INDEX idx_models_name ON models(name);
CREATE INDEX idx_models_status ON models(status);
CREATE INDEX idx_models_is_enabled ON models(is_enabled);
CREATE INDEX idx_models_is_default ON models(is_default);
CREATE INDEX idx_models_supports_embeddings ON models(supports_embeddings);
CREATE INDEX idx_models_last_used ON models(last_used_at);
CREATE INDEX idx_model_configs_model_id ON model_configs(model_id);

-- Triggers to update updated_at timestamp
CREATE TRIGGER update_models_updated_at
    BEFORE UPDATE ON models
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_model_configs_updated_at
    BEFORE UPDATE ON model_configs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger to ensure only one default model
CREATE OR REPLACE FUNCTION ensure_single_default_model()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.is_default = TRUE THEN
        UPDATE models SET is_default = FALSE WHERE id != NEW.id AND is_default = TRUE;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER ensure_single_default_model_trigger
    AFTER UPDATE OF is_default ON models
    FOR EACH ROW
    WHEN (NEW.is_default = TRUE)
    EXECUTE FUNCTION ensure_single_default_model();

-- Trigger to update model last_used_at when used in messages
CREATE OR REPLACE FUNCTION update_model_last_used()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.model IS NOT NULL THEN
        UPDATE models 
        SET last_used_at = NEW.created_at 
        WHERE name = NEW.model;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_model_last_used_trigger
    AFTER INSERT ON messages
    FOR EACH ROW
    EXECUTE FUNCTION update_model_last_used();