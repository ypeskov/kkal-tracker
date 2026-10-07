package server

import (
	"net/http"
	"testing"
	"time"
)

type weightGoalProgress struct {
	TargetWeight        float64 `json:"target_weight"`
	TargetDate          *string `json:"target_date"`
	GoalSetAt           string  `json:"goal_set_at"`
	StartDate           string  `json:"start_date"`
	InitialWeightAtGoal float64 `json:"initial_weight_at_goal"`
	CurrentWeight       float64 `json:"current_weight"`
}

func TestWeightGoalStartDate(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)
	signUp(t, client, smtpServer, "goal-start@example.com")

	for _, entry := range []struct {
		weight float64
		date   string
	}{{82, "2026-01-10"}, {80.5, "2026-01-15"}, {79, "2026-01-20"}} {
		client.expect(http.StatusCreated, http.MethodPost, "/api/weight",
			map[string]any{"weight": entry.weight, "recorded_at": entry.date}, nil)
	}

	// A goal started on a past day takes the weight recorded on that day or before it
	var goal weightGoalProgress
	client.expect(http.StatusOK, http.MethodPut, "/api/profile/goal",
		map[string]any{"target_weight": 70.0, "start_date": "2026-01-17"}, &goal)
	if goal.StartDate != "2026-01-17" || goal.InitialWeightAtGoal != 80.5 {
		t.Fatalf("goal = %+v, want start 2026-01-17 from 80.5", goal)
	}

	// Editing without a start date, or with the same one, keeps the start
	var edited weightGoalProgress
	client.expect(http.StatusOK, http.MethodPut, "/api/profile/goal",
		map[string]any{"target_weight": 71.0, "start_date": "2026-01-17"}, &edited)
	if edited.TargetWeight != 71 || edited.GoalSetAt != goal.GoalSetAt || edited.InitialWeightAtGoal != 80.5 {
		t.Errorf("edited goal = %+v, want target 71 with the original start", edited)
	}

	// Moving the start to another day re-reads the initial weight from the history
	var moved weightGoalProgress
	client.expect(http.StatusOK, http.MethodPut, "/api/profile/goal",
		map[string]any{"target_weight": 71.0, "start_date": "2026-01-20"}, &moved)
	if moved.StartDate != "2026-01-20" || moved.InitialWeightAtGoal != 79 {
		t.Errorf("moved goal = %+v, want start 2026-01-20 from 79", moved)
	}

	// A start before any record falls back to the earliest entry
	var early weightGoalProgress
	client.expect(http.StatusOK, http.MethodPut, "/api/profile/goal",
		map[string]any{"target_weight": 71.0, "start_date": "2026-01-01"}, &early)
	if early.StartDate != "2026-01-01" || early.InitialWeightAtGoal != 82 {
		t.Errorf("early goal = %+v, want start 2026-01-01 from 82", early)
	}

	// Invalid start dates: in the future, on or after the target date
	client.expect(http.StatusBadRequest, http.MethodPut, "/api/profile/goal",
		map[string]any{"target_weight": 71.0, "start_date": "2099-01-01"}, nil)
	today := time.Now().UTC().Format("2006-01-02")
	client.expect(http.StatusBadRequest, http.MethodPut, "/api/profile/goal",
		map[string]any{"target_weight": 71.0, "start_date": today, "target_date": today}, nil)
	client.expect(http.StatusBadRequest, http.MethodPut, "/api/profile/goal",
		map[string]any{"target_weight": 71.0, "start_date": "not-a-date"}, nil)
}

func TestWeightGoalEditKeepsStart(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)
	signUp(t, client, smtpServer, "goal@example.com")

	client.expect(http.StatusCreated, http.MethodPost, "/api/weight",
		map[string]any{"weight": 80.5, "recorded_at": "2026-01-15"}, nil)

	var goal weightGoalProgress
	client.expect(http.StatusOK, http.MethodPut, "/api/profile/goal", map[string]any{"target_weight": 70.0}, &goal)
	if goal.TargetWeight != 70 || goal.InitialWeightAtGoal != 80.5 || goal.GoalSetAt == "" {
		t.Fatalf("new goal = %+v, want target 70 starting from 80.5", goal)
	}

	// The weight changes after the goal was set
	client.expect(http.StatusCreated, http.MethodPost, "/api/weight",
		map[string]any{"weight": 79.0, "recorded_at": "2026-01-20"}, nil)

	// Editing the target keeps the start of the goal: its date and the initial weight
	var edited weightGoalProgress
	client.expect(http.StatusOK, http.MethodPut, "/api/profile/goal",
		map[string]any{"target_weight": 71.5, "target_date": "2099-12-31"}, &edited)
	if edited.TargetWeight != 71.5 || edited.TargetDate == nil {
		t.Errorf("edited goal = %+v, want target 71.5 with a target date", edited)
	}
	if edited.GoalSetAt != goal.GoalSetAt {
		t.Errorf("goal_set_at changed on edit: %s -> %s", goal.GoalSetAt, edited.GoalSetAt)
	}
	if edited.InitialWeightAtGoal != 80.5 {
		t.Errorf("initial_weight_at_goal changed on edit: %v, want 80.5", edited.InitialWeightAtGoal)
	}
	if edited.CurrentWeight != 79 {
		t.Errorf("current_weight = %v, want 79", edited.CurrentWeight)
	}

	// Removing the deadline is an edit too
	var noDeadline weightGoalProgress
	client.expect(http.StatusOK, http.MethodPut, "/api/profile/goal", map[string]any{"target_weight": 71.5}, &noDeadline)
	if noDeadline.TargetDate != nil || noDeadline.GoalSetAt != goal.GoalSetAt || noDeadline.InitialWeightAtGoal != 80.5 {
		t.Errorf("goal after removing the deadline = %+v, want no date and the original start", noDeadline)
	}

	// A goal set after clearing the previous one starts from the current weight
	client.expect(http.StatusNoContent, http.MethodDelete, "/api/profile/goal", nil, nil)
	var fresh weightGoalProgress
	client.expect(http.StatusOK, http.MethodPut, "/api/profile/goal", map[string]any{"target_weight": 70.0}, &fresh)
	if fresh.InitialWeightAtGoal != 79 {
		t.Errorf("initial weight of a new goal = %v, want 79", fresh.InitialWeightAtGoal)
	}
}
