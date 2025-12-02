# TODO

## Features

### 1. llm
Elysia client wrapper for embeddings and agent.

- [ ] Initialize OpenAI client with OpenRouter base URL
- [ ] `Embed(ctx, text string) ([]float64, error)`
- [ ] Agent setup for RAG responses
- [ ] Config: `OPENROUTER_API_KEY`, `EMBED_MODEL`, `CHAT_MODEL`

### 2. vectorstore
Qdrant operations.

- [ ] Implement `qdrant/qdrant.go` client
- [ ] `Upsert(ctx, id string, vector []float64, payload map[string]any) error`
- [ ] `Search(ctx, vector []float64, limit int) ([]SearchResult, error)`
- [ ] Create collection on startup if not exists

### 3. pokemon
PokeAPI client.

- [ ] `GetByID(ctx, id int) (*Pokemon, error)`
- [ ] `GetByName(ctx, name string) (*Pokemon, error)`
- [ ] Return raw JSON for embedding

### 4. indexer
Kafka consumer → fetch → embed → store.

- [ ] Implement Kafka consumer
- [ ] Listen for pokemon IDs on topic
- [ ] Fetch pokemon from API
- [ ] Embed raw JSON
- [ ] Upsert to Qdrant with pokemon metadata
- [ ] Emulate Kafka for local dev (HTTP endpoint or CLI)

### 5. query
RAG HTTP endpoint.

- [ ] `POST /query` - `{ "question": "...", "conversation_id": "..." }`
- [ ] Embed question
- [ ] Search Qdrant for relevant pokemon
- [ ] Build prompt with retrieved context
- [ ] Call agent for response
- [ ] Return answer

### 6. conversation (later)
Redis conversation history.

- [ ] Store conversation turns by conversation_id
- [ ] TTL: 3 minutes
- [ ] Include in RAG prompt context
