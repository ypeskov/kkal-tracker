-- +goose Up
CREATE TABLE ai_providers (
    id VARCHAR(50) PRIMARY KEY,
    display_name VARCHAR(100) NOT NULL,
    model VARCHAR(100) NOT NULL,
    is_active BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Seed default providers (inactive by default until API keys are configured)
INSERT INTO ai_providers (id, display_name, model, is_active) VALUES
    ('openai', 'OpenAI', 'gpt-4o-mini', false),
    ('anthropic', 'Anthropic', 'claude-sonnet-4-20250514', false),
    ('ollama', 'Ollama (Local)', 'llama3', false);

-- +goose Down
DROP TABLE IF EXISTS ai_providers;




