CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL,
    role TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS payments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    amount REAL NOT NULL,
    currency TEXT NOT NULL,
    status TEXT NOT NULL
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
    (200.00, 'USD', 'completed');
