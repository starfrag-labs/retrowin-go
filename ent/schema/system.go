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

// System holds the schema definition for a system (computer/node) in the cluster.
type System struct {
	ent.Schema
}

// Annotations of the System.
func (System) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "systems"},
	}
}

// Mixin of the System.
func (System) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

// Fields of the System.
func (System) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("id"),

		// System name/hostname
		field.String("name").
			MaxLen(64),

		// System description
		field.String("description").
			Optional().
			Nillable().
			MaxLen(255),

		// System status
		field.Enum("status").
			Values("active", "inactive", "maintenance").
			Default("active"),
	}
}

// Indexes of the System.
func (System) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name"),
		index.Fields("status"),
	}
}

// Edges of the System.
func (System) Edges() []ent.Edge {
	return []ent.Edge{
		// System has many inodes
		edge.To("inodes", Inode.Type),

		// System has many users (M2M)
		edge.From("users", User.Type).
			Ref("systems"),
	}
}
