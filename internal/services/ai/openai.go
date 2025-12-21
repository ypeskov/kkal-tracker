package ai

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"text/template"
	"time"

	"ypeskov/kkal-tracker/internal/config"

	"github.com/sashabaranov/go-openai"
)

var (
	systemPromptTmpl *template.Template
	userPromptTmpl   *template.Template
)

func init() {
	var err error
	systemPromptTmpl, err = template.New("system").Parse(SystemPromptTemplate)
	if err != nil {
		panic(fmt.Sprintf("failed to parse system prompt template: %v", err))
	}

	userPromptTmpl, err = template.New("user").Parse(UserPromptTemplate)
	if err != nil {
		panic(fmt.Sprintf("failed to parse user prompt template: %v", err))
	}
}

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

// systemPromptData holds data for system prompt template
type systemPromptData struct {
	Language string
	Age      *int
	Height   *float64
}

// buildSystemPrompt creates the system prompt for AI analysis
func (p *OpenAIProvider) buildSystemPrompt(userCtx UserContext) string {
	lang := userCtx.Language
	if lang == "" {
		lang = config.DefaultLanguageCode
	}

	data := systemPromptData{
		Language: config.GetLanguageName(lang),
		Age:      userCtx.Age,
		Height:   userCtx.Height,
	}

	var buf bytes.Buffer
	if err := systemPromptTmpl.Execute(&buf, data); err != nil {
		p.logger.Error("Failed to execute system prompt template", slog.String("error", err.Error()))
		return ""
	}

	return buf.String()
}

// userPromptData holds data for user prompt template
type userPromptData struct {
	PeriodDays    int
	NutritionData []NutritionDataPoint
	WeightData    []WeightDataPoint
	Query         string
}

// buildUserPrompt creates the user prompt with nutrition and weight data
func (p *OpenAIProvider) buildUserPrompt(req AnalysisRequest) string {
	data := userPromptData{
		PeriodDays:    req.PeriodDays,
		NutritionData: req.NutritionData,
		WeightData:    req.WeightData,
		Query:         req.Query,
	}

	var buf bytes.Buffer
	if err := userPromptTmpl.Execute(&buf, data); err != nil {
		p.logger.Error("Failed to execute user prompt template", slog.String("error", err.Error()))
		return ""
	}

	return buf.String()
}
