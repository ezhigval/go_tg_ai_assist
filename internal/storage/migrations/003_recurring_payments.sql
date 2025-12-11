CREATE TABLE IF NOT EXISTS recurring_payments (
                                                  id SERIAL PRIMARY KEY,
                                                  user_id BIGINT NOT NULL,
                                                  title TEXT NOT NULL,
                                                  amount NUMERIC(14,2) NOT NULL,
    category TEXT,
    period TEXT NOT NULL,
    next_payment DATE NOT NULL,
    created_at TIMESTAMP DEFAULT now()
    );