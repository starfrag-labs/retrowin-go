package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// SystemGroup holds the schema definition for system groups.
type SystemGroup struct {
	ent.Schema
}

// Annotations of the SystemGroup.
func (SystemGroup) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "system_groups"},
	}
}

// Fields of the SystemGroup.
func (SystemGroup) Fields() []ent.Field {
	return []ent.Field{
		field.String("system_id"),

		// Group name
		field.String("name").
			MaxLen(32),

		// Internal group ID within the system
		field.Int("gid").
			Default(0),
	}
}

// Indexes of the SystemGroup.
func (SystemGroup) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("system_id", "name").Unique(),
		index.Fields("system_id", "gid").Unique(),
		index.Fields("system_id"),
	}
}

// Edges of the SystemGroup.
func (SystemGroup) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("system", System.Type).
			Unique().
			Required().
			Field("system_id"),

		// Many-to-many with UserSystem through UserGroup
		edge.From("users", UserSystem.Type).
			Ref("groups"),
	}
}
