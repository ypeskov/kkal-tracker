-- +goose Up
DROP TABLE IF EXISTS ai_providers;

-- +goose Down
CREATE TABLE ai_providers (
    id VARCHAR(50) PRIMARY KEY,
    display_name VARCHAR(100) NOT NULL,
    model VARCHAR(100) NOT NULL,
    provider_type VARCHAR(50) DEFAULT 'openai',
    is_active BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO ai_providers (id, provider_type, display_name, model, is_active) VALUES
    ('openai-gpt-5.2', 'openai', 'GPT-5.2', 'gpt-5.2', false);
