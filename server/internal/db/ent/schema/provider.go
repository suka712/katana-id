package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Provider links a user to an external OAuth identity (Google, GitHub, ...).
type Provider struct {
	ent.Schema
}

func (Provider) Fields() []ent.Field {
	return []ent.Field{
		field.String("provider_name"),
		field.String("provider_account_id"),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (Provider) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("providers").Unique().Required(),
	}
}

func (Provider) Indexes() []ent.Index {
	return []ent.Index{
		// One row per external identity.
		index.Fields("provider_name", "provider_account_id").Unique(),
	}
}
