//go:generate go run -mod=mod entgo.io/ent/cmd/ent generate --feature sql/execquery ./

package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.Int32("id").
			Positive(),
		field.String("username").
			MaxLen(64).
			Unique(),
		field.String("password_hash").
			MaxLen(255),
		field.String("salt").
			MaxLen(64).
			Default(""),
		field.Bool("is_member").
			Default(false),
		field.Int16("member_level").
			Default(0),
		field.Time("member_expire_at").
			Optional().
			Nillable(),
		field.Int32("free_trial_count").
			Default(1),
		field.String("avatar").
			Optional(),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return nil
}
