package skill

import (
	"errors"
	"strings"
	"testing"
	"unicode/utf8"
)

const (
	// developmentURL is the base URL of the development server
	developmentURL = "https://dev-kcal.peskov.info"
	testKey        = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
)

func mustRender(t *testing.T, development bool, baseURL string) *Skill {
	t.Helper()
	rendered, err := Render(Embedded(), development, baseURL)
	if err != nil {
		t.Fatalf("Render(development=%v, %q): %v", development, baseURL, err)
	}
	return rendered
}

func fileContent(t *testing.T, s *Skill, path string) string {
	t.Helper()
	for _, file := range s.Files {
		if file.Path == path {
			return string(file.Content)
		}
	}
	t.Fatalf("skill %s has no file %s", s.Name, path)
	return ""
}

func TestRenderProductionKeepsTheSources(t *testing.T) {
	rendered := mustRender(t, false, ProductionURL)

	if rendered.Name != ProductionName {
		t.Errorf("name = %q, want %q", rendered.Name, ProductionName)
	}
	if len(rendered.Files) != 2 {
		t.Fatalf("files = %d, want SKILL.md and the client only", len(rendered.Files))
	}
	for path, source := range map[string]string{SkillFile: Embedded().Skill, ClientFile: Embedded().Client} {
		if fileContent(t, rendered, path) != source {
			t.Errorf("%s of the production variant differs from the source", path)
		}
	}
}

func TestRenderSubstitutesTheBaseURL(t *testing.T) {
	for _, development := range []bool{false, true} {
		rendered := mustRender(t, development, "https://example.test/")

		skillText := fileContent(t, rendered, SkillFile)
		client := fileContent(t, rendered, ClientFile)
		if !strings.Contains(skillText, "- Base URL: `https://example.test`\n") {
			t.Errorf("development=%v: SKILL.md has no Base URL line with the trimmed URL", development)
		}
		if !strings.Contains(client, `DEFAULT_BASE_URL = "https://example.test"`+"\n") {
			t.Errorf("development=%v: client has no DEFAULT_BASE_URL with the trimmed URL", development)
		}
		if !strings.Contains(client, `API_KEY = "API_KEY_HERE"`+"\n") {
			t.Errorf("development=%v: client lost the API key placeholder", development)
		}
		if strings.Contains(skillText, "`"+ProductionURL+"`") || strings.Contains(client, `"`+ProductionURL+`"`) {
			t.Errorf("development=%v: the production URL is left in the files", development)
		}
	}
}

func TestRenderDevelopmentVariant(t *testing.T) {
	rendered := mustRender(t, true, developmentURL)
	skillText := fileContent(t, rendered, SkillFile)

	if rendered.Name != DevelopmentName {
		t.Errorf("name = %q, want %q", rendered.Name, DevelopmentName)
	}
	for _, part := range []string{
		"---\nname: kkal-tracker-food-log-dev\ndescription: DEV/TEST copy of the Kkal Tracker food logging skill.",
		"# Kkal Tracker: food diary logging (DEV)\n\n> **DEV skill.**",
		`Start every reply with "[DEV]"`,
		"- Base URL: `https://dev-kcal.peskov.info`",
	} {
		if !strings.Contains(skillText, part) {
			t.Errorf("dev SKILL.md does not contain %q", part)
		}
	}
	if client := fileContent(t, rendered, ClientFile); !strings.Contains(client, `DEFAULT_BASE_URL = "https://dev-kcal.peskov.info"`+"\n") {
		t.Error("dev client does not point at the development server")
	}
}

func TestFrontmatterNameOfBothVariants(t *testing.T) {
	for _, development := range []bool{false, true} {
		rendered := mustRender(t, development, ProductionURL)

		name, err := FrontmatterName(fileContent(t, rendered, SkillFile))
		if err != nil {
			t.Fatalf("development=%v: %v", development, err)
		}
		if name != rendered.Name {
			t.Errorf("development=%v: frontmatter name %q differs from the folder name %q", development, name, rendered.Name)
		}
		if err := ValidateName(name); err != nil {
			t.Errorf("development=%v: %v", development, err)
		}
	}
}

// Agent skill uploads reject a description longer than 1024 characters.
func TestDescriptionLengthOfBothVariants(t *testing.T) {
	const maxDescriptionLength = 1024
	for _, development := range []bool{false, true} {
		rendered := mustRender(t, development, ProductionURL)

		var description string
		for line := range strings.SplitSeq(fileContent(t, rendered, SkillFile), "\n") {
			if after, ok := strings.CutPrefix(line, "description: "); ok {
				description = after
				break
			}
		}
		if description == "" {
			t.Fatalf("development=%v: no description in the frontmatter", development)
		}
		if n := utf8.RuneCountInString(description); n > maxDescriptionLength {
			t.Errorf("development=%v: description has %d characters, at most %d allowed", development, n, maxDescriptionLength)
		}
	}
}

func TestValidateName(t *testing.T) {
	for _, name := range []string{ProductionName, DevelopmentName, "a", strings.Repeat("a", 64)} {
		if err := ValidateName(name); err != nil {
			t.Errorf("ValidateName(%q) = %v, want nil", name, err)
		}
	}
	for _, name := range []string{"", strings.Repeat("a", 65), "Kkal", "kkal_food", "kkal food", "-kkal", "kkal-",
		"kkal--food", "claude-food-log", "anthropic-skill"} {
		if err := ValidateName(name); err == nil {
			t.Errorf("ValidateName(%q) = nil, want an error", name)
		}
	}
}

func TestRenderRejectsInvalidBaseURL(t *testing.T) {
	for _, baseURL := range []string{
		"", "/", "kcal.peskov.info", "ftp://kcal.peskov.info", "https://", "https://user:pass@kcal.peskov.info",
		"https://kcal.peskov.info/?a=1", "https://kcal.peskov.info/#x", `https://kcal.peskov.info/"+x`,
		"https://kcal.peskov.info/a b", "https://kcal.peskov.info/\nx",
	} {
		if _, err := Render(Embedded(), false, baseURL); !errors.Is(err, ErrInvalidBaseURL) {
			t.Errorf("Render(%q) error = %v, want ErrInvalidBaseURL", baseURL, err)
		}
	}
}

func TestRenderRequiresEveryAnchorExactlyOnce(t *testing.T) {
	src := Embedded()
	baseURLLine := "- Base URL: `" + ProductionURL + "`"
	defaultURL := `DEFAULT_BASE_URL = "` + ProductionURL + `"`
	keyLine := `API_KEY = "API_KEY_HERE"`

	tests := []struct {
		name        string
		source      Source
		development bool
	}{
		{"no frontmatter", Source{Skill: strings.TrimPrefix(src.Skill, "---\n"), Client: src.Client}, false},
		{"name not on line 2", Source{Skill: strings.Replace(src.Skill, "name: ", "title: ", 1), Client: src.Client}, true},
		{"description not on line 3",
			Source{Skill: strings.Replace(src.Skill, "description: ", "summary: ", 1), Client: src.Client}, true},
		{"empty source", Source{}, false},
		{"title missing", Source{Skill: strings.Replace(src.Skill, title, "# Kkal\n", 1), Client: src.Client}, true},
		{"title twice", Source{Skill: src.Skill + "\n" + title, Client: src.Client}, true},
		{"base URL line missing",
			Source{Skill: strings.Replace(src.Skill, baseURLLine, "- Base URL: elsewhere", 1), Client: src.Client}, false},
		{"base URL line twice", Source{Skill: src.Skill + baseURLLine + "\n", Client: src.Client}, true},
		{"DEFAULT_BASE_URL missing",
			Source{Skill: src.Skill, Client: strings.Replace(src.Client, defaultURL, `DEFAULT_BASE_URL = ""`, 1)}, false},
		{"DEFAULT_BASE_URL twice", Source{Skill: src.Skill, Client: src.Client + defaultURL + "\n"}, true},
		{"API key line missing",
			Source{Skill: src.Skill, Client: strings.Replace(src.Client, keyLine, `API_KEY = ""`, 1)}, false},
		{"API key line twice", Source{Skill: src.Skill, Client: src.Client + keyLine + "\n"}, false},
		{"placeholder outside the key line",
			Source{Skill: src.Skill, Client: src.Client + "# replace API_KEY_HERE\n"}, true},
		{"placeholder only outside the key line", Source{Skill: src.Skill,
			Client: strings.Replace(src.Client, keyLine, `API_KEY = "x"  # API_KEY_HERE`, 1)}, false},
		{"placeholder in SKILL.md", Source{Skill: src.Skill + "Paste the key over API_KEY_HERE.\n", Client: src.Client}, false},
		{"production name changed",
			Source{Skill: strings.Replace(src.Skill, "name: "+ProductionName, "name: Food Log", 1), Client: src.Client}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rendered, err := Render(tt.source, tt.development, ProductionURL)
			if !errors.Is(err, ErrSourceLayout) {
				t.Errorf("error = %v, want ErrSourceLayout", err)
			}
			if rendered != nil {
				t.Error("a skill was returned together with the error")
			}
		})
	}
}

func TestTemplateHoldsTheKeyPlaceholderOnce(t *testing.T) {
	src := Embedded()
	if n := strings.Count(src.Client, APIKeyPlaceholder); n != 1 || !strings.Contains(src.Client, `API_KEY = "API_KEY_HERE"`+"\n") {
		t.Errorf("template client has %d placeholders, want exactly one, in the API_KEY line", n)
	}
	if strings.Contains(src.Skill, APIKeyPlaceholder) {
		t.Error("SKILL.md contains the key placeholder")
	}
}

func TestWithAPIKeyPutsTheKeyIntoTheClient(t *testing.T) {
	for _, development := range []bool{false, true} {
		rendered := mustRender(t, development, "https://example.test")
		withKey, err := rendered.WithAPIKey(testKey)
		if err != nil {
			t.Fatalf("development=%v: WithAPIKey: %v", development, err)
		}

		if withKey.Name != rendered.Name || len(withKey.Files) != len(rendered.Files) {
			t.Fatalf("development=%v: skill with key = %s with %d files, want %s with %d",
				development, withKey.Name, len(withKey.Files), rendered.Name, len(rendered.Files))
		}
		client := fileContent(t, withKey, ClientFile)
		if !strings.Contains(client, `API_KEY = "`+testKey+`"`+"\n") || strings.Contains(client, APIKeyPlaceholder) {
			t.Errorf("development=%v: client does not hold exactly the key in place of the placeholder", development)
		}
		if strings.Replace(client, testKey, APIKeyPlaceholder, 1) != fileContent(t, rendered, ClientFile) {
			t.Errorf("development=%v: the key substitution changed more than the key", development)
		}
		if fileContent(t, withKey, SkillFile) != fileContent(t, rendered, SkillFile) {
			t.Errorf("development=%v: the key substitution changed SKILL.md", development)
		}
		// The rendered skill itself stays a template
		if !strings.Contains(fileContent(t, rendered, ClientFile), `API_KEY = "API_KEY_HERE"`) {
			t.Errorf("development=%v: WithAPIKey changed the rendered skill", development)
		}
	}
}

func TestWithAPIKeyRejectsInvalidKeys(t *testing.T) {
	rendered := mustRender(t, false, ProductionURL)
	for name, key := range map[string]string{
		"empty":     "",
		"63 chars":  testKey[1:],
		"65 chars":  testKey + "0",
		"uppercase": strings.ToUpper(testKey),
		"quote":     `0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcd"+`,
		"newline":   testKey[:32] + "\n" + testKey[33:],
		"non-hex":   strings.Repeat("g", 64),
	} {
		withKey, err := rendered.WithAPIKey(key)
		if !errors.Is(err, ErrInvalidAPIKey) || withKey != nil {
			t.Errorf("%s: WithAPIKey = %v, %v; want ErrInvalidAPIKey", name, withKey, err)
			continue
		}
		if key != "" && strings.Contains(err.Error(), key) {
			t.Errorf("%s: the error contains the key", name)
		}
	}
}

func TestWithAPIKeyRequiresTheKeyLineOnce(t *testing.T) {
	keyLine := `API_KEY = "API_KEY_HERE"`
	for name, client := range map[string]string{
		"key line missing": strings.Replace(Embedded().Client, keyLine, `API_KEY = ""`, 1),
		"key line twice":   Embedded().Client + keyLine + "\n",
	} {
		// A skill built by hand: Render would already reject such a client
		broken := &Skill{Name: ProductionName, Files: []File{{Path: ClientFile, Content: []byte(client)}}}
		withKey, err := broken.WithAPIKey(testKey)
		if !errors.Is(err, ErrSourceLayout) || withKey != nil {
			t.Errorf("%s: WithAPIKey = %v, %v; want ErrSourceLayout", name, withKey, err)
		} else if strings.Contains(err.Error(), testKey) {
			t.Errorf("%s: the error contains the key", name)
		}
	}

	noClient := &Skill{Name: ProductionName, Files: []File{{Path: SkillFile, Content: []byte(Embedded().Skill)}}}
	if _, err := noClient.WithAPIKey(testKey); !errors.Is(err, ErrSourceLayout) {
		t.Errorf("skill without a client: error = %v, want ErrSourceLayout", err)
	}
}
