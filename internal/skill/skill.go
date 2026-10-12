// Package skill renders the food logging agent skill (skills/kkal-tracker-food-log) for a server.
//
// The production variant is the skill as it is, pointed at the server's URL. The development variant is the
// "-dev" copy: another name and description and a DEV banner, so both can be installed in an agent side by side.
// This is the only implementation of the dev transformation: the skill download endpoint and cmd/skillgen,
// which writes the committed skills/kkal-tracker-food-log-dev, both use it.
package skill

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"ypeskov/kkal-tracker/skills"
)

const (
	ProductionName  = "kkal-tracker-food-log"
	DevelopmentName = "kkal-tracker-food-log-dev"

	// ProductionURL is the base URL written in the skill sources
	ProductionURL = "https://kcal.peskov.info"
	// DevelopmentURL is the base URL of the committed dev copy
	DevelopmentURL = "https://dev-kcal.peskov.info"

	// Paths of the skill files inside the skill folder
	SkillFile  = "SKILL.md"
	ClientFile = "scripts/kkal_client.py"

	title = "# Kkal Tracker: food diary logging\n"

	devDescription = `DEV/TEST copy of the Kkal Tracker food logging skill. It writes to the development server, ` +
		`not to the real diary. Use it ONLY when the current message itself names the dev or test ` +
		`tracker — "запиши на дев", "в тестовый трекер", "на тестовый сервер", "log it to dev". The mention is ` +
		`required in every request — an earlier dev request in the same conversation does not carry over. ` +
		`The one exception is editing or deleting an entry that this dev copy created, or an ingredient from ` +
		`its ingredient list, which is always done here. ` +
		`For every other ` +
		`"I ate X" / "запиши еду" request use the regular kkal-tracker-food-log skill instead.`
	devBanner = `> **DEV skill.** Everything below goes to the development server, not to the production diary. ` +
		`Start every reply with "[DEV]" so the user can see where the meal was written.` + "\n"
)

var (
	ErrInvalidBaseURL = errors.New("base URL must be an absolute http(s) URL without credentials, query or quotes")
	ErrSourceLayout   = errors.New("unexpected skill source layout")

	// Agent Skills name rules: lowercase letters, digits and hyphens, at most 64 characters
	validName     = regexp.MustCompile(`^[a-z0-9-]{1,64}$`)
	reservedWords = []string{"anthropic", "claude"}
)

// Source holds the files of the production skill
type Source struct {
	Skill  string // SKILL.md
	Client string // scripts/kkal_client.py
}

// Embedded returns the production skill compiled into the binary
func Embedded() Source {
	return Source{Skill: skills.FoodLogSkill, Client: skills.FoodLogClient}
}

// File is a file of the skill; Path is relative to the skill folder, with forward slashes
type File struct {
	Path    string
	Content []byte
}

// Skill is a rendered skill: its name (also the folder name) and its files.
// It never contains an api_key file.
type Skill struct {
	Name  string
	Files []File
}

// Render builds the production or the development variant of the skill pointed at baseURL.
// Every anchor of the transformation must occur exactly once in the source, otherwise an error is returned.
func Render(src Source, development bool, baseURL string) (*Skill, error) {
	baseURL, err := normalizeBaseURL(baseURL)
	if err != nil {
		return nil, err
	}

	name := ProductionName
	if development {
		name = DevelopmentName
	}

	skillText, err := renderSkillText(src.Skill, development, baseURL)
	if err != nil {
		return nil, err
	}
	client, err := replaceOnce(ClientFile, src.Client,
		`DEFAULT_BASE_URL = "`+ProductionURL+`"`, `DEFAULT_BASE_URL = "`+baseURL+`"`)
	if err != nil {
		return nil, err
	}

	if err := checkName(skillText, name); err != nil {
		return nil, err
	}

	return &Skill{
		Name: name,
		Files: []File{
			{Path: SkillFile, Content: []byte(skillText)},
			{Path: ClientFile, Content: []byte(client)},
		},
	}, nil
}

// WriteDir writes the skill files into dir, creating the folders. Other files in dir, such as api_key, stay untouched.
func (s *Skill) WriteDir(dir string) error {
	for _, file := range s.Files {
		path := filepath.Join(dir, filepath.FromSlash(file.Path))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, file.Content, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func renderSkillText(text string, development bool, baseURL string) (string, error) {
	lines := strings.Split(text, "\n")
	if len(lines) < 3 || lines[0] != "---" || !strings.HasPrefix(lines[1], "name: ") ||
		!strings.HasPrefix(lines[2], "description: ") {
		return "", fmt.Errorf("%w: %s: frontmatter must start with the name and description lines", ErrSourceLayout, SkillFile)
	}

	if development {
		lines[1] = "name: " + DevelopmentName
		lines[2] = "description: " + devDescription
		text = strings.Join(lines, "\n")

		var err error
		text, err = replaceOnce(SkillFile, text, title, strings.TrimSuffix(title, "\n")+" (DEV)\n\n"+devBanner)
		if err != nil {
			return "", err
		}
	}

	return replaceOnce(SkillFile, text, "- Base URL: `"+ProductionURL+"`", "- Base URL: `"+baseURL+"`")
}

// replaceOnce replaces old with new and fails unless old occurs exactly once
func replaceOnce(file, text, old, new string) (string, error) {
	if count := strings.Count(text, old); count != 1 {
		return "", fmt.Errorf("%w: %s: expected exactly one occurrence of %q, found %d", ErrSourceLayout, file, old, count)
	}
	return strings.Replace(text, old, new, 1), nil
}

// normalizeBaseURL drops trailing slashes and rejects URLs that do not fit into the skill files
func normalizeBaseURL(raw string) (string, error) {
	baseURL := strings.TrimRight(raw, "/")
	parsed, err := url.Parse(baseURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" ||
		parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" ||
		strings.ContainsAny(baseURL, "\"'`\\ \t\r\n") {
		return "", ErrInvalidBaseURL
	}
	return baseURL, nil
}

// checkName verifies that the frontmatter name equals the folder name and is a valid skill name
func checkName(skillText, want string) error {
	name, err := FrontmatterName(skillText)
	if err != nil {
		return err
	}
	if name != want {
		return fmt.Errorf("%w: %s: name %q, want %q", ErrSourceLayout, SkillFile, name, want)
	}
	return ValidateName(name)
}

// FrontmatterName returns the name field of a SKILL.md frontmatter
func FrontmatterName(skillText string) (string, error) {
	lines := strings.SplitN(skillText, "\n", 3)
	if len(lines) < 2 || lines[0] != "---" || !strings.HasPrefix(lines[1], "name: ") {
		return "", fmt.Errorf("%w: %s: no name in the frontmatter", ErrSourceLayout, SkillFile)
	}
	return strings.TrimPrefix(lines[1], "name: "), nil
}

// ValidateName checks a skill name against the Agent Skills rules
func ValidateName(name string) error {
	if !validName.MatchString(name) || strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") ||
		strings.Contains(name, "--") {
		return fmt.Errorf("%w: invalid skill name %q", ErrSourceLayout, name)
	}
	for _, word := range reservedWords {
		if strings.Contains(name, word) {
			return fmt.Errorf("%w: skill name %q contains the reserved word %q", ErrSourceLayout, name, word)
		}
	}
	return nil
}
