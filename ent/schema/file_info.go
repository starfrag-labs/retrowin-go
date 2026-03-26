package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// FileInfo holds the schema definition for the FileInfo entity.
type FileInfo struct {
	ent.Schema
}

// Fields of the FileInfo.
func (FileInfo) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("file_id").
			Unique(),
		field.Time("create_date").
			Default(time.Now),
		field.Time("update_date").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.Int64("byte_size").
			Default(0),
	}
}

// Indexes of the FileInfo.
func (FileInfo) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("file_id").Unique(),
	}
}
