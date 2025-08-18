-- Fix embedding dimensions for nomic-embed-text model (768 dimensions)
-- This migration updates all vector columns from 1536 to 768 dimensions

-- Update message_embeddings table
ALTER TABLE message_embeddings ALTER COLUMN embedding TYPE vector(768);

-- Update memory_summaries table
ALTER TABLE memory_summaries ALTER COLUMN embedding TYPE vector(768);

-- Update semantic_topics table
ALTER TABLE semantic_topics ALTER COLUMN embedding TYPE vector(768);

-- Add index for better performance on similarity searches
CREATE INDEX IF NOT EXISTS idx_message_embeddings_vector ON message_embeddings USING ivfflat (embedding vector_cosine_ops);
CREATE INDEX IF NOT EXISTS idx_memory_summaries_vector ON memory_summaries USING ivfflat (embedding vector_cosine_ops);
CREATE INDEX IF NOT EXISTS idx_semantic_topics_vector ON semantic_topics USING ivfflat (embedding vector_cosine_ops);