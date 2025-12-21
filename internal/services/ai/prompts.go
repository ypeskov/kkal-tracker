package ai

import (
	_ "embed"
)

//go:embed prompts/system.txt
var SystemPromptTemplate string

//go:embed prompts/user.txt
var UserPromptTemplate string
