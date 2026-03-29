package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Symlink holds the schema definition for symbolic links.
// Only inodes with file_type=symlink have symlink records.
type Symlink struct {
	ent.Schema
}

// Annotations of the Symlink (table name: symlinks).
func (Symlink) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "symlinks"},
	}
}

// Fields of the Symlink.
func (Symlink) Fields() []ent.Field {
	return []ent.Field{
		// Inode ID this symlink belongs to (one-to-one)
		field.Int64("inode_id").
			Unique(),

		// Target path (the path this symlink points to)
		field.String("target_path").
			MaxLen(4096),
	}
}

// Indexes of the Symlink.
func (Symlink) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("inode_id").Unique(),
	}
}

// Edges of the Symlink.
func (Symlink) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("inode", Inode.Type).
			Unique().
			Required().
			Field("inode_id"),
	}
}
