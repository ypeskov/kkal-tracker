package ai

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"ypeskov/kkal-tracker/internal/config"

	"github.com/sashabaranov/go-openai"
)

// OpenAIProvider implements the Provider interface for OpenAI
type OpenAIProvider struct {
	client       *openai.Client
	logger       *slog.Logger
	useMaxTokens bool
	maxTokens    int
}

// NewOpenAIProvider creates a new OpenAI provider instance
func NewOpenAIProvider(apiKey string, baseURL string, useMaxTokens bool, maxTokens int, logger *slog.Logger) *OpenAIProvider {
	cfg := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		cfg.BaseURL = baseURL
	}

	return &OpenAIProvider{
		client:       openai.NewClientWithConfig(cfg),
		logger:       logger.With("provider", "openai"),
		useMaxTokens: useMaxTokens,
		maxTokens:    maxTokens,
	}
}

// GetProviderName returns the provider identifier
func (p *OpenAIProvider) GetProviderName() string {
	return "openai"
}

// IsAvailable checks if the provider is properly configured
func (p *OpenAIProvider) IsAvailable() bool {
	return p.client != nil
}

// Analyze performs nutrition and weight analysis using OpenAI
func (p *OpenAIProvider) Analyze(ctx context.Context, model string, req AnalysisRequest) (*AnalysisResponse, error) {
	startTime := time.Now()

	systemPrompt := p.buildSystemPrompt(req.UserContext)
	userPrompt := p.buildUserPrompt(req)

	p.logger.Debug("Sending analysis request to OpenAI",
		slog.String("model", model),
		slog.Int("nutrition_points", len(req.NutritionData)),
		slog.Int("weight_points", len(req.WeightData)),
		slog.Bool("use_max_tokens", p.useMaxTokens),
		slog.Int("max_tokens", p.maxTokens))

	chatReq := openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: userPrompt,
			},
		},
		Temperature: 0.7,
	}

	// Use MaxCompletionTokens for newer models (GPT-4.5+, GPT-5.x)
	if p.useMaxTokens && p.maxTokens > 0 {
		chatReq.MaxCompletionTokens = p.maxTokens
	}

	resp, err := p.client.CreateChatCompletion(ctx, chatReq)
	if err != nil {
		p.logger.Error("OpenAI API error", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%w: %v", ErrAnalysisFailed, err)
	}

	if len(resp.Choices) == 0 {
		return nil, ErrAnalysisFailed
	}

	duration := time.Since(startTime).Milliseconds()

	return &AnalysisResponse{
		Analysis:   resp.Choices[0].Message.Content,
		Model:      model,
		TokensUsed: resp.Usage.TotalTokens,
		DurationMs: duration,
	}, nil
}

// buildSystemPrompt creates the system prompt for AI analysis
func (p *OpenAIProvider) buildSystemPrompt(userCtx UserContext) string {
	lang := userCtx.Language
	if lang == "" {
		lang = config.DefaultLanguageCode
	}

	// Determine response language from config
	responseLang := config.GetLanguageName(lang)

	prompt := fmt.Sprintf(`You are a professional nutritionist and health advisor.
Analyze the user's nutrition and weight data to provide personalized insights and recommendations.

Guidelines:
- Respond in %s language
- Be encouraging but honest about areas for improvement
- Provide specific, actionable recommendations
- Consider trends over time, not just individual days
- If weight data is available, correlate it with calorie intake
- Analyze macronutrient balance (proteins, fats, carbohydrates) - not just total calories
- Comment on protein intake adequacy, fat quality distribution, and carbohydrate levels
- Identify any imbalances in the macronutrient ratios
- Keep the analysis concise but comprehensive (2-3 paragraphs max)

IMPORTANT - Output Format:
- Return your response as clean HTML (no markdown)
- Use HTML tags for formatting: <h3> for section headers, <p> for paragraphs, <ul>/<li> for bullet lists, <strong> for emphasis
- Do NOT wrap the response in code blocks or backticks
- Do NOT include <html>, <head>, or <body> tags - just the content HTML
- Example structure: <h3>Section Title</h3><p>Content here...</p><ul><li>Item 1</li><li>Item 2</li></ul>

User profile:`, responseLang)

	if userCtx.Age != nil {
		prompt += fmt.Sprintf("\n- Age: %d years", *userCtx.Age)
	}
	if userCtx.Height != nil {
		prompt += fmt.Sprintf("\n- Height: %.1f cm", *userCtx.Height)
	}

	return prompt
}

// buildUserPrompt creates the user prompt with nutrition and weight data
func (p *OpenAIProvider) buildUserPrompt(req AnalysisRequest) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Please analyze my nutrition and weight data from the last %d days.\n\n", req.PeriodDays))

	// Add nutrition data
	if len(req.NutritionData) > 0 {
		sb.WriteString("## Daily Nutrition Data:\n")
		for _, d := range req.NutritionData {
			sb.WriteString(fmt.Sprintf("### %s: %d kcal total, %.1fg fat, %.1fg carbs, %.1fg protein\n",
				d.Date, d.Calories, d.Fats, d.Carbs, d.Proteins))

			// Add food items for this day
			if len(d.FoodItems) > 0 {
				sb.WriteString("Foods eaten:\n")
				for _, item := range d.FoodItems {
					sb.WriteString(fmt.Sprintf("  - %s (%.0fg, %d kcal)\n", item.Name, item.Weight, item.Calories))
				}
			}
			sb.WriteString("\n")
		}
	}

	// Add weight data
	if len(req.WeightData) > 0 {
		sb.WriteString("## Weight Measurements:\n")
		for _, w := range req.WeightData {
			sb.WriteString(fmt.Sprintf("- %s: %.1f kg\n", w.Date, w.Weight))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("Please provide:\n")
	sb.WriteString("1. Overall assessment of my nutrition habits\n")
	sb.WriteString("2. Analysis of calorie intake trends\n")
	sb.WriteString("3. Macronutrient balance analysis (proteins, fats, carbs ratio and adequacy)\n")
	if len(req.WeightData) > 0 {
		sb.WriteString("4. Weight trend analysis and correlation with nutrition\n")
	}
	sb.WriteString("5. Specific recommendations for improvement\n")

	return sb.String()
}
