package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// DirectoryEntry holds the schema definition for directory entries.
// In Linux, a directory entry (dentry) maps a filename to an inode.
// Multiple entries can point to the same inode (hard links).
type DirectoryEntry struct {
	ent.Schema
}

// Annotations of the DirectoryEntry (table name: directory_entries).
func (DirectoryEntry) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "directory_entries"},
	}
}

// Fields of the DirectoryEntry.
func (DirectoryEntry) Fields() []ent.Field {
	return []ent.Field{
		// Custom int64 ID
		field.Int64("id").
			StorageKey("id"),

		// Parent directory inode ID
		field.Int64("parent_id"),

		// Filename within the parent directory
		field.String("name").
			MaxLen(255),

		// Target inode ID
		field.Int64("child_id"),
	}
}

// Indexes of the DirectoryEntry.
func (DirectoryEntry) Indexes() []ent.Index {
	return []ent.Index{
		// Unique constraint: (parent_id, name) must be unique
		index.Fields("parent_id", "name").Unique(),
		// Index for looking up all entries in a directory
		index.Fields("parent_id"),
		// Index for looking up all names pointing to an inode (hard link count)
		index.Fields("child_id"),
	}
}

// Edges of the DirectoryEntry.
func (DirectoryEntry) Edges() []ent.Edge {
	return []ent.Edge{
		// Edge to parent directory (Inode with type=directory)
		edge.To("parent", Inode.Type).
			Unique().
			Required().
			Field("parent_id"),

		// Edge to target inode
		edge.To("child", Inode.Type).
			Unique().
			Required().
			Field("child_id"),
	}
}
