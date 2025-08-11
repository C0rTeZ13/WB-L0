CREATE TABLE items (
                       id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
                       chrt_id BIGINT NOT NULL,
                       track_number VARCHAR(255) NOT NULL,
                       price INT NOT NULL,
                       rid VARCHAR(255) NOT NULL,
                       name VARCHAR(255) NOT NULL,
                       sale INT NOT NULL,
                       size VARCHAR(50),
                       total_price INT NOT NULL,
                       nm_id BIGINT NOT NULL,
                       brand VARCHAR(255),
                       status INT NOT NULL,
                       created_at TIMESTAMP DEFAULT NOW(),
                       updated_at TIMESTAMP DEFAULT NOW()
);
