package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
)

// FileLink holds the schema definition for the FileLink entity.
type FileLink struct {
	ent.Schema
}

// Mixin of the FileLink.
func (FileLink) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

// Fields of the FileLink.
func (FileLink) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("file_id"),
		field.Int64("target_id"),
	}
}

// Indexes of the FileLink.
func (FileLink) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("file_id", "target_id").Unique(),
		index.Fields("target_id"),
	}
}
