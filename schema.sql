-- Core User Profiles
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- OAuth credentials for GitHub Syncing
CREATE TABLE oauth_tokens (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(20) NOT NULL DEFAULT 'github',
    access_token TEXT NOT NULL,
    refresh_token TEXT,
    expiry TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Registered MCP Repositories
CREATE TABLE repositories (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    repo_url TEXT NOT NULL,
    local_path TEXT NOT NULL UNIQUE,
    current_status VARCHAR(20) DEFAULT 'idle', -- 'idle', 'testing', 'error'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_user_repo UNIQUE(user_id, name)
);