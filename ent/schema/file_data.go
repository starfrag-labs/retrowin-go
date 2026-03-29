package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// FileData holds the schema definition for file content storage.
// Only regular files have file_data records (not directories or symlinks).
type FileData struct {
	ent.Schema
}

// Annotations of the FileData (table name: file_data).
func (FileData) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "file_data"},
	}
}

// Fields of the FileData.
func (FileData) Fields() []ent.Field {
	return []ent.Field{
		// Inode ID this data belongs to (one-to-one)
		field.Int64("inode_id").
			Unique(),

		// Storage type: s3, local, etc.
		field.Enum("storage_type").
			Values("s3", "local").
			Default("s3"),

		// Location in storage (e.g., S3 key, local file path)
		field.String("location").
			MaxLen(1024),

		// Checksum for integrity (e.g., SHA-256)
		field.String("checksum").
			Optional().
			Nillable().
			MaxLen(64),
	}
}

// Indexes of the FileData.
func (FileData) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("inode_id").Unique(),
		index.Fields("storage_type"),
	}
}

// Edges of the FileData.
func (FileData) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("inode", Inode.Type).
			Unique().
			Required().
			Field("inode_id"),
	}
}
