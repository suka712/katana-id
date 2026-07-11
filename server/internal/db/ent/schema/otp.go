package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// OTP is a one-time passcode emailed for passwordless sign-in.
type OTP struct {
	ent.Schema
}

// Annotations pins the table name; without it Ent snake-cases "OTP" to "ot_ps".
func (OTP) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "otps"},
	}
}

func (OTP) Fields() []ent.Field {
	return []ent.Field{
		field.String("email"),
		field.String("code"),
		field.Time("expires_at"),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (OTP) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("email"),
	}
}
