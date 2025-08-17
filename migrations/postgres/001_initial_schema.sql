-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Sessions table
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL DEFAULT 'New Chat',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Messages table
CREATE TABLE messages (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('user', 'assistant', 'system')),
    content TEXT NOT NULL,
    model TEXT,
    tokens_used INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

-- Message embeddings table for semantic search
CREATE TABLE message_embeddings (
    id TEXT PRIMARY KEY,
    message_id TEXT NOT NULL,
    embedding vector(1536), -- Default dimension for many embedding models
    model_used TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE
);

-- Memory summaries table for conversation consolidation
CREATE TABLE memory_summaries (
    id TEXT PRIMARY KEY,
    session_id TEXT,
    summary_type TEXT NOT NULL CHECK (summary_type IN ('conversation', 'topic', 'period', 'global')),
    title TEXT,
    content TEXT NOT NULL,
    embedding vector(1536),
    relevance_score REAL DEFAULT 0.0,
    start_time TIMESTAMP WITH TIME ZONE,
    end_time TIMESTAMP WITH TIME ZONE,
    message_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

-- Memory gaps table for tracking context discontinuities
CREATE TABLE memory_gaps (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    gap_start TIMESTAMP WITH TIME ZONE NOT NULL,
    gap_end TIMESTAMP WITH TIME ZONE NOT NULL,
    context_summary TEXT,
    bridge_content TEXT,
    gap_type TEXT DEFAULT 'temporal' CHECK (gap_type IN ('temporal', 'topical', 'contextual')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

-- Semantic topics table for categorization
CREATE TABLE semantic_topics (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    embedding vector(1536),
    message_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Message-topic relationships
CREATE TABLE message_topics (
    message_id TEXT NOT NULL,
    topic_id TEXT NOT NULL,
    relevance_score REAL DEFAULT 0.0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (message_id, topic_id),
    FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
    FOREIGN KEY (topic_id) REFERENCES semantic_topics(id) ON DELETE CASCADE
);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger to update session updated_at when messages are inserted
CREATE OR REPLACE FUNCTION update_session_on_message_insert()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE sessions
    SET updated_at = CURRENT_TIMESTAMP
    WHERE id = NEW.session_id;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_session_on_message_insert
    AFTER INSERT ON messages
    FOR EACH ROW
    EXECUTE FUNCTION update_session_on_message_insert();

-- Trigger to update memory_summaries updated_at
CREATE TRIGGER update_memory_summaries_updated_at
    BEFORE UPDATE ON memory_summaries
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger to update semantic_topics updated_at
CREATE TRIGGER update_semantic_topics_updated_at
    BEFORE UPDATE ON semantic_topics
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();