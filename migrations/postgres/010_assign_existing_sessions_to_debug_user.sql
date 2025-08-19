-- Assign existing sessions with NULL user_id to the debug user
-- This ensures that existing sessions are visible after authentication is enabled

UPDATE sessions 
SET user_id = 'debug-user-id' 
WHERE user_id IS NULL;

-- Also update any memory summaries that might be orphaned
UPDATE memory_summaries 
SET user_id = 'debug-user-id' 
WHERE user_id IS NULL;

-- Update semantic topics as well
UPDATE semantic_topics 
SET user_id = 'debug-user-id' 
WHERE user_id IS NULL;