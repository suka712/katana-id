package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// User is a KatanaID account, identified by email. Accounts are created lazily
// on first OTP verification or first OAuth sign-in.
type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("username").NotEmpty(),
		field.String("email").Unique().NotEmpty(),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("providers", Provider.Type),
		edge.To("kits", BrandKit.Type),
	}
}
