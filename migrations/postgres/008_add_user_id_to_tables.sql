-- Add user_id column to sessions table
ALTER TABLE sessions ADD COLUMN user_id TEXT;

-- Add user_id column to memory_summaries table (for global summaries)
ALTER TABLE memory_summaries ADD COLUMN user_id TEXT;

-- Add user_id column to semantic_topics table (for user-specific topics)
ALTER TABLE semantic_topics ADD COLUMN user_id TEXT;

-- Add foreign key constraints
ALTER TABLE sessions ADD CONSTRAINT fk_sessions_user_id 
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE memory_summaries ADD CONSTRAINT fk_memory_summaries_user_id 
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE semantic_topics ADD CONSTRAINT fk_semantic_topics_user_id 
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Create indexes for performance
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_memory_summaries_user_id ON memory_summaries(user_id);
CREATE INDEX idx_semantic_topics_user_id ON semantic_topics(user_id);

-- Update existing sessions to have a null user_id (will need manual assignment or cleanup)
-- In production, you might want to create a default user or handle this differently