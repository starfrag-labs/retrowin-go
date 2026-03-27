package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
)

// File holds the schema definition for the File entity.
type File struct {
	ent.Schema
}

// Mixin of the File.
func (File) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

// Fields of the File.
func (File) Fields() []ent.Field {
	return []ent.Field{
		field.String("file_key").
			Unique().
			Immutable(),
		field.Enum("type").
			Values("container", "file"),
		field.String("file_name").
			MaxLen(255),
		field.Int64("owner_id"),
		field.Int64("parent_id").
			Optional().
			Nillable(),
		field.Int64("byte_size").
			Default(0),
		field.Bool("is_system").
			Default(false),
		field.String("system_type").
			Optional().
			Nillable().
			MaxLen(50),
	}
}

// Indexes of the File.
func (File) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("file_key").Unique(),
		index.Fields("owner_id", "parent_id"),
		index.Fields("owner_id", "is_system", "system_type"),
	}
}
