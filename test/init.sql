CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE payments (
    id SERIAL PRIMARY KEY,
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    status VARCHAR(50) NOT NULL
);

INSERT INTO users (email, role) VALUES
    ('alice@example.com', 'admin'),
    ('bob@example.com', 'user'),
    ('charlie@example.com', 'user'),
    ('diana@example.com', 'moderator'),
    ('eve@example.com', 'user');

INSERT INTO payments (amount, currency, status) VALUES
    (99.99, 'USD', 'completed'),
    (49.50, 'EUR', 'pending'),
    (150.00, 'USD', 'completed'),
    (25.00, 'GBP', 'failed'),
    (200.00, 'KRW', 'completed');
