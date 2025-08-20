CREATE TABLE deliveries (
                            id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
                            name VARCHAR(255) NOT NULL,
                            phone VARCHAR(50) NOT NULL,
                            zip VARCHAR(20) NOT NULL,
                            city VARCHAR(100) NOT NULL,
                            address TEXT NOT NULL,
                            region VARCHAR(100) NOT NULL,
                            email VARCHAR(255) NOT NULL,
                            created_at TIMESTAMP DEFAULT NOW(),
                            updated_at TIMESTAMP DEFAULT NOW()
);
