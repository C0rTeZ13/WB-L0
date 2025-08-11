package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"time"
)

// Item holds the schema definition for the Item entity.
type Item struct {
	ent.Schema
}

// Fields of the Item.
func (Item) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Unique().
			Immutable().
			Positive().
			StorageKey("id"),
		field.Int64("chrt_id").
			Positive(),
		field.String("track_number").
			MaxLen(255).
			NotEmpty(),
		field.Int("price").
			Positive(),
		field.String("rid").
			MaxLen(255).
			NotEmpty(),
		field.String("name").
			MaxLen(255).
			NotEmpty(),
		field.Int("sale").
			NonNegative(),
		field.String("size").
			MaxLen(50).
			Optional().
			Nillable(),
		field.Int("total_price").
			Positive(),
		field.Int64("nm_id").
			Positive(),
		field.String("brand").
			MaxLen(255).
			Optional().
			Nillable(),
		field.Int("status"),
		field.Time("created_at").
			Default(func() time.Time { return time.Now() }).
			Immutable(),
		field.Time("updated_at").
			Default(func() time.Time { return time.Now() }).
			UpdateDefault(func() time.Time { return time.Now() }),

		field.Int("order_id").
			Optional().
			Nillable(),
	}
}

// Edges of the Item.
func (Item) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("order", Order.Type).
			Ref("items").
			Field("order_id").
			Unique(),
	}
}
