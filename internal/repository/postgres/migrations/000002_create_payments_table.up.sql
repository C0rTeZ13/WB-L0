CREATE TABLE payments (
                          id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
                          "transaction" VARCHAR(255) NOT NULL,
                          request_id VARCHAR(255),
                          currency VARCHAR(10) NOT NULL,
                          provider VARCHAR(100) NOT NULL,
                          amount INT NOT NULL,
                          payment_dt BIGINT NOT NULL,
                          bank VARCHAR(100) NOT NULL,
                          delivery_cost INT NOT NULL,
                          goods_total INT NOT NULL,
                          custom_fee INT NOT NULL,
                          created_at TIMESTAMP DEFAULT NOW(),
                          updated_at TIMESTAMP DEFAULT NOW()
);
