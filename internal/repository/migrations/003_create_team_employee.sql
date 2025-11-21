CREATE TABLE IF NOT EXISTS teams (
    id SERIAL PRIMARY KEY,
    team_id VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    parent_id VARCHAR(100),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);


CREATE TABLE IF NOT EXISTS employees (
    id BIGINT PRIMARY KEY,
    username VARCHAR(200),
    email VARCHAR(200),
    updated_at TIMESTAMP DEFAULT NOW()
);
    