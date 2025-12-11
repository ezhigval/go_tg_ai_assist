CREATE TABLE IF NOT EXISTS users (
                                     id BIGSERIAL PRIMARY KEY,
                                     user_id BIGINT UNIQUE NOT NULL,
                                     chat_id BIGINT NOT NULL,
                                     last_message TEXT,
                                     last_seen TIMESTAMP DEFAULT now()
    );

CREATE TABLE IF NOT EXISTS user_states (
                                           user_id BIGINT PRIMARY KEY,
                                           state_json TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS todos (
                                     id SERIAL PRIMARY KEY,
                                     user_id BIGINT NOT NULL,
                                     title TEXT NOT NULL,
                                     description TEXT,
                                     due_date TIMESTAMP,
                                     status TEXT NOT NULL DEFAULT 'pending',
                                     created_at TIMESTAMP DEFAULT now()
    );

CREATE TABLE IF NOT EXISTS credits (
                                       id SERIAL PRIMARY KEY,
                                       user_id BIGINT NOT NULL,
                                       title TEXT NOT NULL,
                                       principal NUMERIC(14,2) NOT NULL,
    rate NUMERIC(6,2) NOT NULL,
    months INT NOT NULL,
    created_at TIMESTAMP DEFAULT now()
    );

CREATE TABLE IF NOT EXISTS finance_entries (
                                               id SERIAL PRIMARY KEY,
                                               user_id BIGINT NOT NULL,
                                               amount NUMERIC(14,2) NOT NULL,
    category TEXT,
    type TEXT NOT NULL,
    note TEXT,
    created_at TIMESTAMP DEFAULT now()
    );