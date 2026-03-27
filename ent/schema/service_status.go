package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ServiceStatus holds the schema definition for the ServiceStatus entity.
type ServiceStatus struct {
	ent.Schema
}

// Fields of the ServiceStatus.
func (ServiceStatus) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id").
			Unique(),
		field.Bool("available").
			Default(true),
		field.Time("join_date").
			Default(time.Now),
		field.Time("update_date").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Indexes of the ServiceStatus.
func (ServiceStatus) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id").Unique(),
	}
}
