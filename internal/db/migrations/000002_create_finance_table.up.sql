CREATE TABLE IF NOT EXISTS finances (
    finance_id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    type VARCHAR(50),  -- 'income' or 'expense'
    source VARCHAR(255),
    amount DECIMAL(10, 2),
    date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    description TEXT,
    FOREIGN KEY (user_id) REFERENCES users (user_id)
);
