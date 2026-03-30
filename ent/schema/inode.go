package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
)

// Inode holds the schema definition for the Inode entity.
// In Linux, an inode contains file metadata but NOT the filename.
// Filenames are stored in directory entries.
type Inode struct {
	ent.Schema
}

// Annotations of the Inode (table name: inodes).
func (Inode) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "inodes"},
	}
}

// Mixin of the Inode.
func (Inode) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

// Fields of the Inode.
func (Inode) Fields() []ent.Field {
	return []ent.Field{
		// Custom int64 ID (inode number)
		field.Int64("id").
			StorageKey("id"),

		// System this inode belongs to
		field.String("system_id"),

		field.Int16("mode"),

		// Owner user ID
		field.Int64("uid").
			Default(0),

		// Owner group ID
		field.Int64("gid").
			Default(0),

		// File size in bytes
		field.Int64("size").
			Default(0),

		// Hard link count
		field.Int("link_count").
			Default(1),

		field.Int16("flags"),

		// Accessed timestamp
		field.Time("atime"),

		// Modified timestamp
		field.Time("mtime"),

		// Changed timestamp
		field.Time("ctime"),

		// Content
		field.Bytes("content").
			Optional(),
	}
}

// Indexes of the Inode.
func (Inode) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("system_id"),
	}
}

// Edges of the Inode.
func (Inode) Edges() []ent.Edge {
	return []ent.Edge{
		// Inode belongs to a system
		edge.From("system", System.Type).
			Ref("inodes").
			Field("system_id").
			Required().
			Unique(),
	}
}
