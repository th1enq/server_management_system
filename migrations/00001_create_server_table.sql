-- +goose Up
CREATE TYPE server_status AS ENUM ('ON', 'OFF');

CREATE TABLE servers (
    id SERIAL PRIMARY KEY,
    server_id VARCHAR(255) NOT NULL UNIQUE,
    server_name VARCHAR(255) NOT NULL UNIQUE,
    status server_status DEFAULT 'OFF',
    ipv4 VARCHAR(45),
    description TEXT,
    location TEXT,
    os TEXT,
    cpu INTEGER,
    ram INTEGER,
    disk INTEGER,
    created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_servers_server_id ON servers (server_id);
CREATE INDEX idx_servers_name ON servers (server_name);

-- +goose Down
DROP TABLE IF EXISTS servers;
DROP TYPE IF EXISTS server_status;
