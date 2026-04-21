// Package knowledge provides the KnowledgeBase + RAG primitives used by
// Agent runs. Vector storage is pluggable: pgvector in-DB or external
// (Weaviate, Qdrant, Milvus).
package knowledge

import "context"

// Document is a raw input to be chunked + embedded.
type Document struct {
	ID     string
	Source string
	Text   string
	Meta   map[string]any
}

// Chunk is a stored segment retrievable by similarity.
type Chunk struct {
	ID     string
	DocID  string
	Seq    int
	Text   string
	Score  float64
	Vector []float32
}

// Store persists chunks and executes similarity search.
type Store interface {
	Upsert(ctx context.Context, kbID uint, docs []Document) error
	Retrieve(ctx context.Context, kbID uint, query string, k int) ([]Chunk, error)
	Delete(ctx context.Context, kbID uint, docID string) error
}

// Embedder turns raw text into vectors. Typically backed by an OpenAI or
// Jina provider reached through the relay layer itself.
type Embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
}
