-- +goose Up

-- Add provider_type column to distinguish provider type from unique ID
ALTER TABLE ai_providers ADD COLUMN provider_type VARCHAR(50) DEFAULT 'openai';

-- Update existing providers with their provider_type
UPDATE ai_providers SET provider_type = 'openai' WHERE id = 'openai';
UPDATE ai_providers SET provider_type = 'anthropic' WHERE id = 'anthropic';
UPDATE ai_providers SET provider_type = 'ollama' WHERE id = 'ollama';

-- Replace existing openai entry with GPT-4.5
UPDATE ai_providers SET id = 'openai-gpt-4.5', display_name = 'GPT-4.5', model = 'gpt-4.5' WHERE id = 'openai';

-- Insert GPT-5.2 (December 2025 - latest)
INSERT INTO ai_providers (id, provider_type, display_name, model, is_active) VALUES
    ('openai-gpt-5.2', 'openai', 'GPT-5.2', 'gpt-5.2', false);

-- +goose Down

-- Remove GPT-5.2
DELETE FROM ai_providers WHERE id = 'openai-gpt-5.2';

-- Restore original openai entry
UPDATE ai_providers SET id = 'openai', display_name = 'OpenAI', model = 'gpt-4o-mini' WHERE id = 'openai-gpt-4.5';

-- Remove provider_type column
ALTER TABLE ai_providers DROP COLUMN provider_type;

