// Package skills embeds the sources of the production food logging skill into the binary.
//
// The files are listed one by one on purpose: a directory embed would also pick up the
// git-ignored api_key file, which holds a real key on development machines.
package skills

import _ "embed"

// FoodLogSkill is kkal-tracker-food-log/SKILL.md
//
//go:embed kkal-tracker-food-log/SKILL.md
var FoodLogSkill string

// FoodLogClient is kkal-tracker-food-log/scripts/kkal_client.py
//
//go:embed kkal-tracker-food-log/scripts/kkal_client.py
var FoodLogClient string
