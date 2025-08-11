package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"time"
)

// Delivery holds the schema definition for the Delivery entity.
type Delivery struct {
	ent.Schema
}

// Fields of the Delivery.
func (Delivery) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Unique().
			Immutable().
			Positive().
			StorageKey("id"),
		field.String("name").
			MaxLen(255).
			NotEmpty(),
		field.String("phone").
			MaxLen(50).
			NotEmpty(),
		field.String("zip").
			MaxLen(20).
			NotEmpty(),
		field.String("city").
			MaxLen(100).
			NotEmpty(),
		field.String("address").
			NotEmpty().
			StorageKey("address").
			SchemaType(map[string]string{
				dialect.SQLite: "TEXT",
			}),
		field.String("region").
			MaxLen(100).
			NotEmpty(),
		field.String("email").
			MaxLen(255).
			NotEmpty(),
		field.Time("created_at").
			Default(func() time.Time {
				return time.Now()
			}).
			Immutable(),
		field.Time("updated_at").
			Default(func() time.Time {
				return time.Now()
			}).
			UpdateDefault(func() time.Time {
				return time.Now()
			}),
	}
}

// Edges of the Delivery.
func (Delivery) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("orders", Order.Type),
	}
}
