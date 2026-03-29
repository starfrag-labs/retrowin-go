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
		field.Int64("system_id").
			Optional().
			Nillable(),

		// File type: regular file (-), directory (d), symlink (l), etc.
		field.Enum("file_type").
			Values("regular", "directory", "symlink", "block", "char", "socket", "fifo").
			Default("regular"),

		// File size in bytes
		field.Int64("byte_size").
			Default(0),

		// Owner user ID (external uid string)
		field.String("owner_uid").
			MaxLen(64),

		// Owner group ID (external gid string)
		field.String("owner_gid").
			MaxLen(64),

		// Permissions: owner (e.g., "rwx")
		field.String("perm_owner").
			Default("rwx").
			MaxLen(3),

		// Permissions: group (e.g., "r-x")
		field.String("perm_group").
			Default("r-x").
			MaxLen(3),

		// Permissions: others (e.g., "r--")
		field.String("perm_others").
			Default("r--").
			MaxLen(3),

		// Hard link count
		field.Int16("link_count").
			Default(1),

		// Accessed timestamp
		field.Time("accessed_at").
			Optional().
			Nillable(),

		// System file markers
		field.Bool("is_system").
			Default(false),
		field.String("system_type").
			Optional().
			Nillable().
			MaxLen(50),
	}
}

// Indexes of the Inode.
func (Inode) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("system_id"),
		index.Fields("owner_uid"),
		index.Fields("owner_uid", "is_system", "system_type"),
	}
}

// Edges of the Inode.
func (Inode) Edges() []ent.Edge {
	return []ent.Edge{
		// Inode belongs to a system
		edge.From("system", System.Type).
			Ref("inodes").
			Field("system_id").
			Unique(),

		// Inode has directory entries (as child)
		edge.To("entries", DirectoryEntry.Type),
	}
}
