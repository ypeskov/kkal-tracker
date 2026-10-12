package skill

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoSkillsDir is the skills/ folder of the repository, relative to this package
var repoSkillsDir = filepath.Join("..", "..", "skills")

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
		if !strings.Contains(client, `API_KEY = "XXXXXXXXX"`+"\n") {
			t.Errorf("development=%v: client lost the API key placeholder", development)
		}
		if strings.Contains(skillText, "`"+ProductionURL+"`") || strings.Contains(client, `"`+ProductionURL+`"`) {
			t.Errorf("development=%v: the production URL is left in the files", development)
		}
	}
}

func TestRenderDevelopmentVariant(t *testing.T) {
	rendered := mustRender(t, true, DevelopmentURL)
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

// The committed dev copy must be the generator output: edit the production skill and run `make skill-dev`
func TestCommittedDevCopyIsUpToDate(t *testing.T) {
	rendered := mustRender(t, true, DevelopmentURL)

	for _, file := range rendered.Files {
		committed, err := os.ReadFile(filepath.Join(repoSkillsDir, DevelopmentName, filepath.FromSlash(file.Path)))
		if err != nil {
			t.Fatalf("failed to read the committed dev copy: %v", err)
		}
		if string(committed) != string(file.Content) {
			t.Errorf("skills/%s/%s is out of date: run `make skill-dev`", DevelopmentName, file.Path)
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

func TestWriteDirKeepsOtherFiles(t *testing.T) {
	dir := t.TempDir()
	const keyFile = "local key\n"
	if err := os.WriteFile(filepath.Join(dir, "api_key"), []byte(keyFile), 0o600); err != nil {
		t.Fatal(err)
	}

	rendered := mustRender(t, true, "http://192.0.2.10:8080")
	if err := rendered.WriteDir(dir); err != nil {
		t.Fatalf("WriteDir: %v", err)
	}

	if content, _ := os.ReadFile(filepath.Join(dir, "api_key")); string(content) != keyFile {
		t.Errorf("api_key was changed: %q", content)
	}
	for _, file := range rendered.Files {
		written, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(file.Path)))
		if err != nil || string(written) != string(file.Content) {
			t.Errorf("%s was not written correctly: %v", file.Path, err)
		}
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 3 {
		t.Errorf("dir has %d entries, want api_key, SKILL.md and scripts/", len(entries))
	}
}
