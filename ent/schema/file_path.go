package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// FilePath holds the schema definition for the FilePath entity.
type FilePath struct {
	ent.Schema
}

// Fields of the FilePath.
func (FilePath) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("file_id").
			Unique(),
		field.JSON("path", []int64{}).
			Optional(),
	}
}

// Indexes of the FilePath.
func (FilePath) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("file_id").Unique(),
	}
}
