package skills

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// The embed directives must name single files: a directory or a pattern would pick up stray files (a git-ignored api_key)
func TestEmbedListsFilesExplicitly(t *testing.T) {
	source, err := os.ReadFile("embed.go")
	if err != nil {
		t.Fatalf("failed to read embed.go: %v", err)
	}

	want := []string{"kkal-tracker-food-log/SKILL.md", "kkal-tracker-food-log/scripts/kkal_client.py"}
	var got []string
	for _, match := range regexp.MustCompile(`(?m)^//go:embed (.*)$`).FindAllStringSubmatch(string(source), -1) {
		got = append(got, strings.Fields(match[1])...)
	}

	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Errorf("embedded paths = %q, want exactly %q", got, want)
	}
}

func TestEmbeddedFilesAreTheSources(t *testing.T) {
	for path, embedded := range map[string]string{
		"kkal-tracker-food-log/SKILL.md":               FoodLogSkill,
		"kkal-tracker-food-log/scripts/kkal_client.py": FoodLogClient,
	} {
		onDisk, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read %s: %v", path, err)
		}
		if embedded == "" || embedded != string(onDisk) {
			t.Errorf("embedded %s differs from the file on disk", path)
		}
	}
}
