-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TYPE server_status AS ENUM ('ON', 'OFF', 'UNDEFINED');

CREATE TABLE servers (
    id SERIAL PRIMARY KEY,
    server_id VARCHAR(255) NOT NULL UNIQUE,
    server_name VARCHAR(255) NOT NULL UNIQUE,
    status server_status DEFAULT 'OFF',
    ipv4 VARCHAR(45),
    description TEXT,
    location TEXT,
    os TEXT,
    interval_time BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


CREATE INDEX idx_servers_server_id ON servers (server_id);
CREATE INDEX idx_servers_name ON servers (server_name);

CREATE TABLE outbox (
    id VARCHAR(100) NOT NULL,
    message BYTEA NOT NULL,
    state INT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    lock_id VARCHAR(100) NULL,
    locked_at TIMESTAMP NULL,
    processed_at TIMESTAMP NULL,
    number_of_attempts INT NOT NULL,
    last_attempt_at TIMESTAMP NULL,
    error VARCHAR(1000) NULL
);

-- +goose Down
DROP TABLE IF EXISTS servers;
DROP TYPE IF EXISTS server_status;
DROP TABLE IF EXISTS outbox;