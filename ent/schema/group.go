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

// Group holds the schema definition for the Group entity.
// Groups operate within a system.
type Group struct {
	ent.Schema
}

// Annotations of the Group.
func (Group) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "groups"},
	}
}

// Mixin of the Group.
func (Group) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

// Fields of the Group.
func (Group) Fields() []ent.Field {
	return []ent.Field{
		// Custom int64 ID
		field.Int64("id").
			StorageKey("id"),

		// System this group belongs to
		field.Int64("system_id"),

		// External group ID (e.g., UUID) - unique within system
		field.String("gid").
			MaxLen(64),

		// Group name - unique within system
		field.String("groupname").
			MaxLen(32),
	}
}

// Indexes of the Group.
func (Group) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("system_id"),
		index.Fields("gid"),
		index.Fields("groupname"),
		// Unique: (system_id, gid)
		index.Fields("system_id", "gid").Unique(),
		// Unique: (system_id, groupname)
		index.Fields("system_id", "groupname").Unique(),
	}
}

// Edges of the Group.
func (Group) Edges() []ent.Edge {
	return []ent.Edge{
		// Group belongs to a system
		edge.To("system", System.Type).
			Unique().
			Required().
			Field("system_id"),

		// Group has many users (M2M)
		edge.From("users", User.Type).
			Ref("groups"),
	}
}
