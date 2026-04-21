package model

import "gorm.io/gorm"

// KnowledgeBase groups documents that an agent may retrieve from.
type KnowledgeBase struct {
	gorm.Model
	OrgID       uint   `gorm:"index"`
	Name        string `gorm:"type:varchar(128)"`
	Embedder    string `gorm:"type:varchar(128)"`
	VectorStore string `gorm:"type:varchar(64)"` // pgvector | qdrant | weaviate | milvus
	Status      string `gorm:"type:varchar(32);default:'ready'"`
}

// KBDocument is a source file / URL / API response ingested into a KB.
type KBDocument struct {
	gorm.Model
	KBID   uint   `gorm:"index"`
	Source string `gorm:"type:varchar(512)"`
	Hash   string `gorm:"type:varchar(128);uniqueIndex:ux_kb_hash,priority:2"`
	KBHashOrg uint `gorm:"uniqueIndex:ux_kb_hash,priority:1"`
	Chunks int
}

// KBChunk is one retrievable segment. Embedding is stored as bytea (or
// pgvector vector type on PostgreSQL via AutoMigrate override).
type KBChunk struct {
	ID        uint   `gorm:"primaryKey"`
	DocID     uint   `gorm:"index"`
	Seq       int
	Text      string `gorm:"type:text"`
	Embedding []byte `gorm:"type:bytea"`
}
