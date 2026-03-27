package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// TempFile holds the schema definition for the TempFile entity.
type TempFile struct {
	ent.Schema
}

// Fields of the TempFile.
func (TempFile) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("owner_id").
			Comment("ID of the user who owns this temp file"),
		field.String("file_key").
			Unique().
			Comment("UUID key for the temp file"),
		field.String("file_name").
			MaxLen(255).
			Comment("Name of the temp file"),
		field.Int64("byte_size").
			Default(0).
			Comment("Size of the temp file in bytes"),
		field.Time("create_date").
			Default(time.Now).
			Comment("Date when the temp file was created"),
		field.Int64("parent_id").
			Optional().
			Nillable().
			Comment("ID of the parent container"),
		field.Time("expires_at").
			Default(func() time.Time {
				return time.Now().Add(24 * time.Hour)
			}).
			Comment("Expiration time for the temp file"),
	}
}

// Indexes of the TempFile.
func (TempFile) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("file_key").Unique(),
		index.Fields("owner_id"),
		index.Fields("expires_at"),
	}
}

// Edges of the TempFile.
func (TempFile) Edges() []ent.Edge {
	return nil
}
