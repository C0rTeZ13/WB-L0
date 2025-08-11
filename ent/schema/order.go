package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"time"
)

// Order holds the schema definition for the Order entity.
type Order struct {
	ent.Schema
}

// Fields of the Order.
func (Order) Fields() []ent.Field {
	return []ent.Field{
		field.String("order_uid").
			NotEmpty().
			Unique(),
		field.String("track_number").
			NotEmpty(),
		field.String("entry").
			NotEmpty(),
		field.String("locale").
			Optional().
			Nillable(),
		field.String("internal_signature").
			Optional().
			Nillable(),
		field.String("customer_id").
			NotEmpty(),
		field.String("delivery_service").
			NotEmpty(),
		field.String("shardkey").
			NotEmpty(),
		field.Int("sm_id").
			Optional().
			Nillable(),
		field.Time("date_created").
			Optional().
			Nillable(),
		field.String("oof_shard").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),

		field.Int("delivery_id").
			Optional().
			Nillable(),
		field.Int("payment_id").
			Optional().
			Nillable(),
	}
}

// Edges of the Order.
func (Order) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("delivery", Delivery.Type).
			Ref("orders").
			Field("delivery_id").
			Unique().
			StructTag(`json:"delivery,omitempty"`),
		edge.From("payment", Payment.Type).
			Ref("orders").
			Field("payment_id").
			Unique().
			StructTag(`json:"payment,omitempty"`),
		edge.To("items", Item.Type),
	}
}
