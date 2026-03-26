package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
)

// FileRole holds the schema definition for the FileRole entity.
type FileRole struct {
	ent.Schema
}

// Mixin of the FileRole.
func (FileRole) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

// Fields of the FileRole.
func (FileRole) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id"),
		field.Int64("file_id"),
		field.JSON("roles", []string{}).
			Default([]string{}),
	}
}

// Indexes of the FileRole.
func (FileRole) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "file_id").Unique(),
		index.Fields("file_id"),
	}
}
