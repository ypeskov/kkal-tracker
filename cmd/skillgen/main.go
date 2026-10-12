// Command skillgen regenerates skills/kkal-tracker-food-log-dev from the production skill.
//
// The dev skill is the production skill pointed at the development server, with the dev name, description and
// banner (internal/skill). The production skill is compiled in, so `go run` always uses the current files.
//
// Usage (from the repository root): go run ./cmd/skillgen [http://HOST:PORT]   or   make skill-dev [URL=...]
// Edit the production skill only, then regenerate. The api_key file of the dev skill (git-ignored) is left
// untouched; the API key in the repository is always the XXXXXXXXX placeholder.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"ypeskov/kkal-tracker/internal/skill"
)

func main() {
	dir := flag.String("dir", filepath.Join("skills", skill.DevelopmentName), "output directory of the dev skill")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: go run ./cmd/skillgen [-dir DIR] [http://HOST:PORT]\n"+
			"Regenerates the dev skill from the production one (default URL %s).\n", skill.DevelopmentURL)
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() > 1 {
		flag.Usage()
		os.Exit(2)
	}
	baseURL := skill.DevelopmentURL
	if flag.NArg() == 1 {
		baseURL = flag.Arg(0)
	}

	// Refuse to create a skills/ tree outside the repository when started from another directory
	if info, err := os.Stat(filepath.Dir(*dir)); err != nil || !info.IsDir() {
		fail(fmt.Errorf("%s does not exist: run from the repository root or pass -dir", filepath.Dir(*dir)))
	}

	rendered, err := skill.Render(skill.Embedded(), true, baseURL)
	if err != nil {
		fail(err)
	}
	if err := rendered.WriteDir(*dir); err != nil {
		fail(err)
	}

	fmt.Printf("Updated %s\n", *dir)
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "skillgen: %v\n", err)
	os.Exit(1)
}
