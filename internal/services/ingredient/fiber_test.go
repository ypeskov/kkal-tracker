package ingredient

import (
	"testing"
)

const fiberLimitMessage = "fiber must be between 0 and 100 (grams per 100 g)"

func fiberOf(t *testing.T, repo *fakeIngredientRepo, id int) any {
	t.Helper()
	got := repo.ingredients[id]
	if got == nil {
		t.Fatalf("ingredient %d does not exist", id)
	}
	return floatValue(got.Fiber)
}

func TestPatchIngredientFiber(t *testing.T) {
	service, repo := newTestService()
	id := repo.insert(user, "Хлеб", 250, ptr(1.0), ptr(50.0), ptr(8.0))

	patch := func(req PatchIngredientRequest) error {
		req.IngredientID, req.UserID = id, user
		_, err := service.PatchIngredient(&req)
		return err
	}

	// Only the fibre changes
	if err := patch(PatchIngredientRequest{Fiber: ptr(2.5)}); err != nil {
		t.Fatalf("PatchIngredient: %v", err)
	}
	assertIngredient(t, repo, id, "Хлеб", 250, 1.0, 50.0, 8.0)
	if got := fiberOf(t, repo, id); got != 2.5 {
		t.Errorf("fiber = %v, want 2.5", got)
	}

	// Another field keeps it
	if err := patch(PatchIngredientRequest{KcalPer100g: ptr(260.0), Name: ptr("Хлеб ржаной")}); err != nil {
		t.Fatalf("PatchIngredient: %v", err)
	}
	if got := fiberOf(t, repo, id); got != 2.5 {
		t.Errorf("fiber after another change = %v, want 2.5", got)
	}

	// The limits are inclusive and 0 is a value
	for _, fiber := range []float64{100, 0} {
		if err := patch(PatchIngredientRequest{Fiber: ptr(fiber)}); err != nil {
			t.Fatalf("PatchIngredient with fiber %g: %v", fiber, err)
		}
		if got := fiberOf(t, repo, id); got != fiber {
			t.Errorf("fiber = %v, want %g", got, fiber)
		}
	}

	// Out of range: nothing changes, also not the valid fields of the same request
	for _, req := range []PatchIngredientRequest{
		{Fiber: ptr(-0.01)}, {Fiber: ptr(100.01)}, {Fiber: ptr(150.0), KcalPer100g: ptr(300.0)},
	} {
		assertValidationError(t, patch(req), fiberLimitMessage)
	}
	assertIngredient(t, repo, id, "Хлеб ржаной", 260, 1.0, 50.0, 8.0)
	if got := fiberOf(t, repo, id); got != 0.0 {
		t.Errorf("fiber after rejected changes = %v, want 0", got)
	}

	// The order of the checks is fats, carbs, proteins, fiber
	assertValidationError(t, patch(PatchIngredientRequest{Proteins: ptr(101.0), Fiber: ptr(101.0)}), "proteins must be between")

	// Nothing sent: the message lists the fibre
	assertValidationError(t, patch(PatchIngredientRequest{}), "nothing to change", "proteins or fiber")
}

func TestAddIngredientFiber(t *testing.T) {
	service, repo := newTestService()

	for name, tt := range map[string]struct {
		fiber *float64
		want  any
	}{
		"Овсянка": {ptr(10.6), 10.6}, "Мясо": {ptr(0.0), 0.0}, "Отруби": {ptr(100.0), 100.0}, "Соус": {nil, nil},
	} {
		got, err := service.AddIngredient(&AddIngredientRequest{UserID: user, Name: ptr(name), KcalPer100g: ptr(100.0), Fiber: tt.fiber})
		if err != nil {
			t.Fatalf("AddIngredient %s: %v", name, err)
		}
		if floatValue(got.Fiber) != tt.want || fiberOf(t, repo, got.ID) != tt.want {
			t.Errorf("%s: fiber = %v, stored %v; want %v", name, floatValue(got.Fiber), fiberOf(t, repo, got.ID), tt.want)
		}
	}

	before := len(repo.ingredients)
	for _, fiber := range []float64{-1, -0.01, 100.01, 150} {
		_, err := service.AddIngredient(addRequest(func(req *AddIngredientRequest) { req.Fiber = ptr(fiber) }))
		assertValidationError(t, err, fiberLimitMessage)
	}
	if len(repo.ingredients) != before {
		t.Errorf("ingredients = %d, want %d: a rejected ingredient was stored", len(repo.ingredients), before)
	}
}

// The website API checks the fibre against both limits, while its macros keep having no upper limit
func TestWebsiteIngredientFiber(t *testing.T) {
	service, repo := newTestService()

	created, err := service.CreateIngredient(&CreateIngredientRequest{
		UserID: user, Name: "Отруби", KcalPer100g: 200, Fats: ptr(120.0), Fiber: ptr(100.0)})
	if err != nil {
		t.Fatalf("CreateIngredient: %v", err)
	}
	assertIngredient(t, repo, created.ID, "Отруби", 200, 120.0, nil, nil)
	if got := fiberOf(t, repo, created.ID); got != 100.0 {
		t.Errorf("fiber = %v, want 100", got)
	}

	for _, fiber := range []float64{-1, 100.01, 101} {
		_, err := service.CreateIngredient(&CreateIngredientRequest{UserID: user, Name: "Плохой", KcalPer100g: 100, Fiber: ptr(fiber)})
		assertValidationError(t, err, fiberLimitMessage)

		_, err = service.UpdateIngredient(&UpdateIngredientRequest{
			IngredientID: created.ID, UserID: user, Name: "Другой", KcalPer100g: 300, Fiber: ptr(fiber)})
		assertValidationError(t, err, fiberLimitMessage)
	}
	if len(repo.ingredients) != 1 {
		t.Errorf("ingredients = %d, want 1", len(repo.ingredients))
	}
	assertIngredient(t, repo, created.ID, "Отруби", 200, 120.0, nil, nil)

	// Update replaces all values: 0 is stored, a fibre that is not sent is cleared like the macros
	update := func(fiber *float64) {
		t.Helper()
		if _, err := service.UpdateIngredient(&UpdateIngredientRequest{
			IngredientID: created.ID, UserID: user, Name: "Отруби", KcalPer100g: 200, Fiber: fiber}); err != nil {
			t.Fatalf("UpdateIngredient: %v", err)
		}
	}
	update(ptr(0.0))
	if got := fiberOf(t, repo, created.ID); got != 0.0 {
		t.Errorf("fiber after an update with 0 = %v, want 0", got)
	}
	update(nil)
	if got := fiberOf(t, repo, created.ID); got != nil {
		t.Errorf("fiber after an update without it = %v, want unknown", got)
	}
}
