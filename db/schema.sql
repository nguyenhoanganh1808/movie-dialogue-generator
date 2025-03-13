CREATE TABLE characters (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    traits JSONB NOT NULL
);

CREATE TABLE dialogues (
    id SERIAL PRIMARY KEY,
    scenario TEXT NOT NULL,
    characters JSONB NOT NULL,
    content JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE reference_dialogues (
    id SERIAL PRIMARY KEY,
    source VARCHAR(255) NOT NULL,
    characters JSONB NOT NULL,
    content TEXT NOT NULL,
    tags TEXT[] NOT NULL
);