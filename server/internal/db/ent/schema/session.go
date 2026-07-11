package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Session is a server-side login session; its token is stored in an HttpOnly
// cookie on the client.
type Session struct {
	ent.Schema
}

func (Session) Fields() []ent.Field {
	return []ent.Field{
		field.String("token").Unique().Immutable().DefaultFunc(uuid.NewString),
		field.String("email"),
		field.Time("expires_at"),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (Session) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("email"),
	}
}
