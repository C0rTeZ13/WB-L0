package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Payment holds the schema definition for the Payment entity.
type Payment struct {
	ent.Schema
}

// Fields of the Payment.
func (Payment) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Unique().
			Immutable().
			Positive().
			StorageKey("id"),
		field.String("transaction").
			MaxLen(255).
			NotEmpty().
			StorageKey("transaction"),
		field.String("request_id").
			MaxLen(255).
			Optional().
			Nillable(),
		field.String("currency").
			MaxLen(10).
			NotEmpty(),
		field.String("provider").
			MaxLen(100).
			NotEmpty(),
		field.Int("amount").
			Positive(),
		field.Int64("payment_dt").
			Positive(),
		field.String("bank").
			MaxLen(100).
			NotEmpty(),
		field.Int("delivery_cost").
			NonNegative(),
		field.Int("goods_total").
			NonNegative(),
		field.Int("custom_fee").
			NonNegative(),
		field.Time("created_at").
			Default(func() time.Time { return time.Now() }).
			Immutable(),
		field.Time("updated_at").
			Default(func() time.Time { return time.Now() }).
			UpdateDefault(func() time.Time { return time.Now() }),
	}
}

// Edges of the Payment.
func (Payment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("orders", Order.Type),
	}
}
