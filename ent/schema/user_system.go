package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// UserSystem holds the schema definition for the user-system relationship.
type UserSystem struct {
	ent.Schema
}

// Annotations of the UserSystem.
func (UserSystem) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "user_systems"},
	}
}

// Fields of the UserSystem.
func (UserSystem) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id"),
		field.String("system_id"),

		// Internal user ID within the system
		field.Int("uid").
			Default(0),

		// Primary group ID
		field.Int("gid").
			Default(0),

		// Username for display
		field.String("username").
			Unique().
			MaxLen(32),
	}
}

// Indexes of the UserSystem.
func (UserSystem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "system_id").Unique(),
		index.Fields("system_id", "uid").Unique(),
		index.Fields("system_id", "username").Unique(),
		index.Fields("system_id"),
	}
}

// Edges of the UserSystem.
func (UserSystem) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Unique().
			Required().
			Field("user_id"),
		edge.To("system", System.Type).
			Unique().
			Required().
			Field("system_id"),
		// Many-to-many relationship with SystemGroup through UserGroup
		edge.To("groups", SystemGroup.Type).
			Through("user_groups", UserGroup.Type),
	}
}
