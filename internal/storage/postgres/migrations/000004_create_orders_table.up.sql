CREATE TABLE orders (
                        id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
                        order_uid TEXT NOT NULL UNIQUE,
                        track_number TEXT NOT NULL UNIQUE,
                        entry TEXT NOT NULL,
                        delivery_id BIGINT,
                        payment_id BIGINT,
                        created_at TIMESTAMP DEFAULT NOW(),
                        updated_at TIMESTAMP DEFAULT NOW(),
                        CONSTRAINT fk_delivery FOREIGN KEY (delivery_id) REFERENCES deliveries(id) ON DELETE SET NULL,
                        CONSTRAINT fk_payment FOREIGN KEY (payment_id) REFERENCES payments(id) ON DELETE SET NULL
);

ALTER TABLE items
    ADD COLUMN order_id BIGINT,
    ADD CONSTRAINT fk_order FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE;
