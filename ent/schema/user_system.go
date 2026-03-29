package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
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
		field.Int64("user_id"),
		field.Int64("system_id"),
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
	}
}
