// Package skills embeds the template of the food logging skill into the binary.
//
// The files are listed one by one on purpose: a directory embed would also pick up stray files, such as
// a git-ignored api_key file of an old setup that may hold a real key.
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
