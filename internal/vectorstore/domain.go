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

type FilterOp string

const (
	FilterAND FilterOp = "AND"
	FilterOR  FilterOp = "OR"
	FilterNOT FilterOp = "NOT"
)

type StringFilter struct {
	Field string
	Value string
	Op    FilterOp
}

type IntFilter struct {
	Field string
	Value int64
	Op    FilterOp
}

type BoolFilter struct {
	Field string
	Value bool
	Op    FilterOp
}

type Filter struct {
	StringFilters []StringFilter
	IntFilters    []IntFilter
	BoolFilters   []BoolFilter
}
