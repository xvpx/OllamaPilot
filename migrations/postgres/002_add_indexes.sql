-- Performance indexes for existing tables
CREATE INDEX idx_messages_session_id ON messages(session_id);
CREATE INDEX idx_messages_created_at ON messages(created_at);
CREATE INDEX idx_messages_role ON messages(role);
CREATE INDEX idx_sessions_updated_at ON sessions(updated_at);

-- Vector similarity search indexes
CREATE INDEX idx_message_embeddings_vector ON message_embeddings USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
CREATE INDEX idx_memory_summaries_vector ON memory_summaries USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
CREATE INDEX idx_semantic_topics_vector ON semantic_topics USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- Indexes for semantic features
CREATE INDEX idx_message_embeddings_message_id ON message_embeddings(message_id);
CREATE INDEX idx_message_embeddings_model ON message_embeddings(model_used);
CREATE INDEX idx_memory_summaries_session_id ON memory_summaries(session_id);
CREATE INDEX idx_memory_summaries_type ON memory_summaries(summary_type);
CREATE INDEX idx_memory_summaries_relevance ON memory_summaries(relevance_score DESC);
CREATE INDEX idx_memory_summaries_time_range ON memory_summaries(start_time, end_time);
CREATE INDEX idx_memory_gaps_session_id ON memory_gaps(session_id);
CREATE INDEX idx_memory_gaps_time_range ON memory_gaps(gap_start, gap_end);
CREATE INDEX idx_memory_gaps_type ON memory_gaps(gap_type);
CREATE INDEX idx_message_topics_message_id ON message_topics(message_id);
CREATE INDEX idx_message_topics_topic_id ON message_topics(topic_id);
CREATE INDEX idx_message_topics_relevance ON message_topics(relevance_score DESC);
CREATE INDEX idx_semantic_topics_name ON semantic_topics(name);
CREATE INDEX idx_semantic_topics_message_count ON semantic_topics(message_count DESC);