package vectorstore

// Point represents a vector point to store.
type Point struct {
	ID      string
	Vector  []float32
	Payload map[string]any
}

// SearchResult represents a single search result with score.
type SearchResult struct {
	ID      string
	Score   float32
	Payload map[string]any
}
