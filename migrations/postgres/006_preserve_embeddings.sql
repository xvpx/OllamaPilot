-- Preserve embeddings when messages are deleted
-- This migration restructures the message_embeddings table to be independent of messages

-- First, add columns to store essential message data directly in embeddings table
ALTER TABLE message_embeddings 
ADD COLUMN IF NOT EXISTS session_id TEXT,
ADD COLUMN IF NOT EXISTS role TEXT,
ADD COLUMN IF NOT EXISTS content TEXT,
ADD COLUMN IF NOT EXISTS message_created_at TIMESTAMP WITH TIME ZONE;

-- Populate the new columns with data from existing messages
UPDATE message_embeddings 
SET 
    session_id = m.session_id,
    role = m.role,
    content = m.content,
    message_created_at = m.created_at
FROM messages m 
WHERE message_embeddings.message_id = m.id;

-- Drop the foreign key constraint that causes CASCADE DELETE
ALTER TABLE message_embeddings 
DROP CONSTRAINT IF EXISTS message_embeddings_message_id_fkey;

-- Make the new columns NOT NULL (after populating them)
ALTER TABLE message_embeddings 
ALTER COLUMN session_id SET NOT NULL,
ALTER COLUMN role SET NOT NULL,
ALTER COLUMN content SET NOT NULL,
ALTER COLUMN message_created_at SET NOT NULL;

-- Add check constraint for role (only if it doesn't exist)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'check_role'
        AND conrelid = 'message_embeddings'::regclass
    ) THEN
        ALTER TABLE message_embeddings
        ADD CONSTRAINT check_role CHECK (role IN ('user', 'assistant', 'system'));
    END IF;
END $$;

-- Add index for better performance on session-based queries
CREATE INDEX IF NOT EXISTS idx_message_embeddings_session_id ON message_embeddings(session_id);
CREATE INDEX IF NOT EXISTS idx_message_embeddings_role ON message_embeddings(role);
CREATE INDEX IF NOT EXISTS idx_message_embeddings_created_at ON message_embeddings(message_created_at);

-- Add index for content search (useful for debugging and admin queries)
CREATE INDEX IF NOT EXISTS idx_message_embeddings_content_gin ON message_embeddings USING gin(to_tsvector('english', content));

-- Optional: Add a comment explaining the design decision
COMMENT ON TABLE message_embeddings IS 'Embeddings are preserved independently of messages to maintain semantic memory across session deletions. Contains denormalized message data for efficient querying.';