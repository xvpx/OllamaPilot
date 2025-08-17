# PRD: Modular Dockerized Ollama Frontend

## 1. Overview
We are building a **Go-based, modular frontend** for **Ollama** to support advanced LLM workflows.  
The system will handle memory, RAG, web search, code execution, and support modular extensions such as **MCP servers**.  

The solution will be **dockerized**, self-hostable, and optimized for **low-latency concurrency** while maintaining security and extensibility.

---

## 2. Goals & Objectives
- **Unified Frontend:** Expose a consistent API for chat, RAG, memory, search, and execution.
- **Streaming:** Support real-time streaming of LLM tokens via SSE/WebSockets.
- **Modularity:** Allow new tools (MCP or custom) to be added with minimal code changes.
- **Persistence:** Provide short- and long-term memory using SQL + vector storage.
- **Security:** Strong sandboxing for untrusted code execution.
- **Scalability:** Support both development and production with PostgreSQL and pgvector.

---

## 3. Key Features

### 3.1 Chat & Orchestration
- API endpoint for initiating chats with Ollama models.
- Stream responses token-by-token.
- Handle tool calls inline (memory, search, RAG, exec).

### 3.2 Memory
- **Short-term:** In-session context tracking.
- **Long-term:** Persisted memory in PostgreSQL.
- **Embeddings:** Periodic summarization & embeddings stored in Qdrant/pgvector.
- **API:** CRUD for user/system memories.

### 3.3 Retrieval-Augmented Generation (RAG)
- **Ingest:** PDFs, HTML, Markdown → chunk → embed with Ollama → store in vector DB.
- **Query:** Top-k similarity search + metadata filters.
- **Citations:** Return text with source attribution.

### 3.4 Web Search
- **Abstraction Layer:** `SearchProvider` interface.
- **Default:** Integrate with a meta-search API.
- **Pluggable:** Add Bing/Brave/Tavily/custom crawlers later.

### 3.5 Code Execution
- **Isolation:** Runs in separate sandbox container/VM/WASI.
- **Resource Limits:** CPU, RAM, runtime limits.
- **Streaming Logs:** User sees real-time execution output.
- **API:** Submit code, poll/stream results, fetch artifacts.

### 3.6 MCP Server Integration
- Adapter layer for **MCP JSON-RPC protocol**.
- Dynamically register tools exposed by MCP servers.
- Expose unified tool registry through `/v1/tools`.

---

## 4. System Architecture

### 4.1 Containers
- **api**: Go HTTP server (chat orchestration, tool routing).
- **ollama**: LLM runtime.
- **vector-db**: Qdrant or Postgres+pgvector.
- **sql-db**: PostgreSQL for all environments.
- **exec-sandbox**: Code execution workers.
- **optional**: redis (for queues/caching).

### 4.2 API Endpoints
- `/v1/chat` → chat with tool streaming.
- `/v1/memory` → CRUD memory.
- `/v1/embed` → embeddings via Ollama.
- `/v1/rag/ingest` → upload documents.
- `/v1/rag/query` → retrieve context.
- `/v1/search` → federated web search.
- `/v1/exec` → run code in sandbox.
- `/v1/tools` → list/enable MCP + custom tools.

---

## 5. Non-Goals
- Not building an entire IDE or UI—this is API-first.
- Not providing pre-built MCP servers; only adapters.
- Not competing with enterprise LLM ops platforms (LangChain/LlamaIndex)—lean, modular, and ops-friendly.

---

## 6. Technical Requirements

### 6.1 Language & Frameworks
- **Go** 1.22+
- HTTP router: `chi` or `gorilla/mux`
- DB: `pgx`/`sqlc`
- Vector DB: Qdrant Go client / pgvector via SQL
- Config: `godotenv` or `viper`
- Observability: `otel`, Prometheus, pprof

### 6.2 Persistence
- PostgreSQL (all environments)

### 6.3 Security
- API keys per user/project
- Strict sandboxing for code execution
- No host mounts or privileged containers
- Provenance metadata for RAG snippets

---

## 7. Milestones

1. **MVP (Core Chat + Ollama)**
   - `/v1/chat` streaming endpoint
   - Memory (in-PostgreSQL)  
2. **RAG**
   - Embeddings via Ollama  
   - Qdrant/pgvector integration  
3. **Web Search**
   - Stub provider  
   - API endpoint  
4. **Code Execution**
   - Sandbox container runner  
   - Streaming logs  
5. **MCP Integration**
   - Register and call MCP tools  
   - Tool registry API  
6. **Production Hardening**
   - Postgres support  
   - Auth, rate limits, observability  

---

## 8. Success Metrics
- **Latency:** <100ms overhead beyond Ollama.
- **Throughput:** Handle 100 concurrent chat sessions.
- **Memory recall:** >80% success on recall benchmarks.
- **RAG accuracy:** Top-k retrieval returns relevant passages 90%+ of the time.
- **Sandbox safety:** No code escape incidents in fuzz testing.
