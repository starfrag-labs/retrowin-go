package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// UserGroup holds the schema for user-group membership (through table).
type UserGroup struct {
	ent.Schema
}

// Annotations of the UserGroup.
func (UserGroup) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "user_groups"},
	}
}

// Fields of the UserGroup.
func (UserGroup) Fields() []ent.Field {
	return []ent.Field{
		field.Int("user_system_id"),
		field.Int("system_group_id"),
	}
}

// Indexes of the UserGroup.
func (UserGroup) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_system_id", "system_group_id").Unique(),
	}
}

// Edges of the UserGroup.
func (UserGroup) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user_system", UserSystem.Type).
			Unique().
			Required().
			Field("user_system_id"),

		edge.To("system_group", SystemGroup.Type).
			Unique().
			Required().
			Field("system_group_id"),
	}
}
