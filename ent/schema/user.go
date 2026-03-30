package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Annotations of the User.
func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "users"},
	}
}

// Mixin of the User.
func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("id"),

		// Username for display
		field.String("username").
			Unique().
			MaxLen(32),

		// OIDC provider info
		field.String("provider").
			MaxLen(32),
		field.String("provider_id").
			MaxLen(255),

		// Join date
		field.Time("join_date").
			Default(time.Now),
	}
}

// Indexes of the User.
func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("provider", "provider_id").Unique(),
		index.Fields("username"),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		// User has access to many systems (M2M through user_systems)
		edge.To("systems", System.Type).
			Through("user_systems", UserSystem.Type),
	}
}
