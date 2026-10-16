package ingredient

import (
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"maps"
	"slices"
	"strings"
	"testing"

	"ypeskov/kkal-tracker/internal/models"
	"ypeskov/kkal-tracker/internal/repositories"
)

// zeroWidthSpace is an invisible character that must not make two names different
const zeroWidthSpace = string(rune(0x200b))

var errUniqueName = errors.New("constraint failed: UNIQUE constraint failed: user_ingredients.user_id, user_ingredients.name")

// fakeIngredientRepo keeps ingredients in memory and enforces the exact (user_id, name) unique index of the database.
// Methods the service does not use come from the embedded nil interface and panic when called.
type fakeIngredientRepo struct {
	repositories.IngredientRepository

	ingredients map[int]*models.UserIngredient
	nextID      int
	// beforeWrite runs right before a create or update, e.g. to simulate a concurrent request
	beforeWrite func()
	// writeErr makes the next create or update fail
	writeErr error
	listErr  error

	// diary holds the usage statistics of food names in the diary of user, matched by the exact name like the database
	diary    map[string]models.UserIngredientUsage
	usageErr error
	// beforeUsage runs right before the usage statistics are read, e.g. to simulate a concurrent delete
	beforeUsage func()
}

func newFakeRepo() *fakeIngredientRepo {
	return &fakeIngredientRepo{ingredients: make(map[int]*models.UserIngredient), nextID: 1}
}

func (r *fakeIngredientRepo) insert(userID int, name string, kcalPer100g float64, fats, carbs, proteins *float64) int {
	id := r.nextID
	r.nextID++
	r.ingredients[id] = &models.UserIngredient{
		ID: id, UserID: userID, Name: name, KcalPer100g: kcalPer100g, Fats: fats, Carbs: carbs, Proteins: proteins,
	}
	return id
}

func (r *fakeIngredientRepo) nameTaken(userID, exceptID int, name string) bool {
	for _, ingredient := range r.ingredients {
		if ingredient.UserID == userID && ingredient.ID != exceptID && ingredient.Name == name {
			return true
		}
	}
	return false
}

func (r *fakeIngredientRepo) startWrite() error {
	if r.beforeWrite != nil {
		r.beforeWrite()
	}
	err := r.writeErr
	r.writeErr = nil
	return err
}

func (r *fakeIngredientRepo) get(userID, ingredientID int) *models.UserIngredient {
	ingredient, ok := r.ingredients[ingredientID]
	if !ok || ingredient.UserID != userID {
		return nil
	}
	copied := *ingredient
	return &copied
}

func (r *fakeIngredientRepo) GetAllUserIngredients(userID int) ([]*models.UserIngredient, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	var result []*models.UserIngredient
	for _, id := range slices.Sorted(maps.Keys(r.ingredients)) {
		if ingredient := r.get(userID, id); ingredient != nil {
			result = append(result, ingredient)
		}
	}
	return result, nil
}

func (r *fakeIngredientRepo) GetUserIngredientsWithUsage(userID int) ([]*models.UserIngredientUsage, error) {
	if r.beforeUsage != nil {
		r.beforeUsage()
	}
	if r.usageErr != nil {
		return nil, r.usageErr
	}
	ingredients, _ := r.GetAllUserIngredients(userID)
	result := make([]*models.UserIngredientUsage, len(ingredients))
	for i, ingredient := range ingredients {
		row := &models.UserIngredientUsage{UserIngredient: *ingredient}
		if used, ok := r.diary[ingredient.Name]; ok && userID == user {
			row.TimesUsed, row.LastUsed = used.TimesUsed, used.LastUsed
		}
		result[i] = row
	}
	return result, nil
}

func (r *fakeIngredientRepo) GetUserIngredientByID(userID, ingredientID int) (*models.UserIngredient, error) {
	if ingredient := r.get(userID, ingredientID); ingredient != nil {
		return ingredient, nil
	}
	return nil, sql.ErrNoRows
}

func (r *fakeIngredientRepo) CreateUserIngredient(userID int, name string, kcalPer100g float64,
	fats, carbs, proteins *float64) (*models.UserIngredient, error) {
	if err := r.startWrite(); err != nil {
		return nil, err
	}
	if r.nameTaken(userID, 0, name) {
		return nil, errUniqueName
	}
	return r.get(userID, r.insert(userID, name, kcalPer100g, fats, carbs, proteins)), nil
}

func (r *fakeIngredientRepo) UpdateUserIngredient(userID, ingredientID int, name string, kcalPer100g float64,
	fats, carbs, proteins *float64) (*models.UserIngredient, error) {
	if err := r.startWrite(); err != nil {
		return nil, err
	}
	if r.nameTaken(userID, ingredientID, name) {
		return nil, errUniqueName
	}
	if ingredient := r.get(userID, ingredientID); ingredient != nil {
		r.ingredients[ingredientID] = &models.UserIngredient{
			ID: ingredientID, UserID: userID, Name: name, KcalPer100g: kcalPer100g, Fats: fats, Carbs: carbs, Proteins: proteins,
		}
	}
	return r.GetUserIngredientByID(userID, ingredientID)
}

const (
	user      = 1
	otherUser = 2
)

func ptr[T any](v T) *T { return &v }

func newTestService() (*Service, *fakeIngredientRepo) {
	repo := newFakeRepo()
	return New(repo, slog.New(slog.NewTextHandler(io.Discard, nil))), repo
}

func floatValue(v *float64) any {
	if v == nil {
		return nil
	}
	return *v
}

// assertIngredient compares the stored state of an ingredient
func assertIngredient(t *testing.T, repo *fakeIngredientRepo, id int, name string, kcal float64, fats, carbs, proteins any) {
	t.Helper()
	got := repo.ingredients[id]
	if got == nil {
		t.Fatalf("ingredient %d does not exist", id)
	}
	if got.Name != name || got.KcalPer100g != kcal ||
		floatValue(got.Fats) != fats || floatValue(got.Carbs) != carbs || floatValue(got.Proteins) != proteins {
		t.Errorf("ingredient %d = %q %g kcal, fats %v, carbs %v, proteins %v; want %q %g kcal, fats %v, carbs %v, proteins %v",
			id, got.Name, got.KcalPer100g, floatValue(got.Fats), floatValue(got.Carbs), floatValue(got.Proteins),
			name, kcal, fats, carbs, proteins)
	}
}

func assertDuplicate(t *testing.T, err error, existingID int) {
	t.Helper()
	var duplicateErr *DuplicateIngredientError
	if !errors.As(err, &duplicateErr) {
		t.Fatalf("error = %v, want a duplicate ingredient error", err)
	}
	if duplicateErr.Existing == nil || duplicateErr.Existing.ID != existingID {
		t.Errorf("existing ingredient = %+v, want ID %d", duplicateErr.Existing, existingID)
	}
	if duplicateErr.Error() != "an ingredient with this name already exists" {
		t.Errorf("message = %q", duplicateErr.Error())
	}
}

func assertValidationError(t *testing.T, err error, wantParts ...string) {
	t.Helper()
	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("error = %v, want a validation error", err)
	}
	for _, part := range wantParts {
		if !strings.Contains(validationErr.Message, part) {
			t.Errorf("message %q does not contain %q", validationErr.Message, part)
		}
	}
}

func TestPatchIngredientMergesSentFields(t *testing.T) {
	service, repo := newTestService()
	breadID := repo.insert(user, "Хлеб", 250, ptr(1.0), ptr(50.0), ptr(8.0))

	// Only the calories change, the name and the nutrients stay
	got, err := service.PatchIngredient(&PatchIngredientRequest{IngredientID: breadID, UserID: user, KcalPer100g: ptr(260.0)})
	if err != nil {
		t.Fatalf("PatchIngredient: %v", err)
	}
	if got.ID != breadID || got.Name != "Хлеб" || got.KcalPer100g != 260 {
		t.Errorf("returned ingredient = %+v", got)
	}
	assertIngredient(t, repo, breadID, "Хлеб", 260, 1.0, 50.0, 8.0)

	// One nutrient
	if _, err := service.PatchIngredient(&PatchIngredientRequest{IngredientID: breadID, UserID: user, Carbs: ptr(48.0)}); err != nil {
		t.Fatalf("PatchIngredient: %v", err)
	}
	assertIngredient(t, repo, breadID, "Хлеб", 260, 1.0, 48.0, 8.0)

	// All five at once
	if _, err := service.PatchIngredient(&PatchIngredientRequest{
		IngredientID: breadID, UserID: user,
		Name: ptr("Хлеб ржаной"), KcalPer100g: ptr(210.0), Fats: ptr(1.2), Carbs: ptr(40.0), Proteins: ptr(8.5),
	}); err != nil {
		t.Fatalf("PatchIngredient: %v", err)
	}
	assertIngredient(t, repo, breadID, "Хлеб ржаной", 210, 1.2, 40.0, 8.5)

	// Unknown nutrients stay unknown when they are not sent
	soupID := repo.insert(user, "Суп", 60, nil, nil, nil)
	if _, err := service.PatchIngredient(&PatchIngredientRequest{IngredientID: soupID, UserID: user, Fats: ptr(2.0)}); err != nil {
		t.Fatalf("PatchIngredient: %v", err)
	}
	assertIngredient(t, repo, soupID, "Суп", 60, 2.0, nil, nil)
}

func TestPatchIngredientLimits(t *testing.T) {
	service, repo := newTestService()
	id := repo.insert(user, "Хлеб", 250, ptr(1.0), ptr(50.0), ptr(8.0))

	valid := []PatchIngredientRequest{
		{KcalPer100g: ptr(0.0)}, {KcalPer100g: ptr(900.0)},
		{Fats: ptr(0.0)}, {Fats: ptr(100.0)},
		{Carbs: ptr(0.0)}, {Carbs: ptr(100.0)},
		{Proteins: ptr(0.0)}, {Proteins: ptr(100.0)},
	}
	for _, req := range valid {
		req.IngredientID, req.UserID = id, user
		if _, err := service.PatchIngredient(&req); err != nil {
			t.Errorf("%+v: %v", req, err)
		}
	}

	repo.ingredients[id].KcalPer100g, repo.ingredients[id].Fats = 250, ptr(1.0)
	repo.ingredients[id].Carbs, repo.ingredients[id].Proteins = ptr(50.0), ptr(8.0)
	invalid := []struct {
		req       PatchIngredientRequest
		wantParts []string
	}{
		{PatchIngredientRequest{KcalPer100g: ptr(-1.0)}, []string{"kcal_per_100g", "900"}},
		{PatchIngredientRequest{KcalPer100g: ptr(-0.1)}, []string{"kcal_per_100g", "900"}},
		{PatchIngredientRequest{KcalPer100g: ptr(900.1)}, []string{"kcal_per_100g", "900"}},
		{PatchIngredientRequest{Fats: ptr(-0.1)}, []string{"fats", "100"}},
		{PatchIngredientRequest{Fats: ptr(100.1)}, []string{"fats", "100"}},
		{PatchIngredientRequest{Carbs: ptr(-0.1)}, []string{"carbs", "100"}},
		{PatchIngredientRequest{Carbs: ptr(100.1)}, []string{"carbs", "100"}},
		{PatchIngredientRequest{Proteins: ptr(-0.1)}, []string{"proteins", "100"}},
		{PatchIngredientRequest{Proteins: ptr(100.1)}, []string{"proteins", "100"}},
		// A valid value does not let an invalid one through
		{PatchIngredientRequest{Name: ptr("Батон"), KcalPer100g: ptr(1000.0)}, []string{"kcal_per_100g"}},
	}
	for _, tt := range invalid {
		tt.req.IngredientID, tt.req.UserID = id, user
		_, err := service.PatchIngredient(&tt.req)
		assertValidationError(t, err, tt.wantParts...)
	}
	assertIngredient(t, repo, id, "Хлеб", 250, 1.0, 50.0, 8.0)
}

func TestPatchIngredientName(t *testing.T) {
	service, repo := newTestService()
	id := repo.insert(user, "Хлеб", 250, nil, nil, nil)

	got, err := service.PatchIngredient(&PatchIngredientRequest{
		IngredientID: id, UserID: user, Name: ptr("  Хлеб" + zeroWidthSpace + "   ржаной "),
	})
	if err != nil {
		t.Fatalf("PatchIngredient: %v", err)
	}
	if got.Name != "Хлеб ржаной" {
		t.Errorf("name = %q, want the cleaned name", got.Name)
	}

	for _, name := range []string{"", "   ", zeroWidthSpace + " \t" + zeroWidthSpace} {
		_, err := service.PatchIngredient(&PatchIngredientRequest{IngredientID: id, UserID: user, Name: ptr(name)})
		assertValidationError(t, err, "name")
	}
	_, err = service.PatchIngredient(&PatchIngredientRequest{IngredientID: id, UserID: user, Name: ptr(strings.Repeat("я", 201))})
	assertValidationError(t, err, "name", "200")
	assertIngredient(t, repo, id, "Хлеб ржаной", 250, nil, nil, nil)

	// 200 characters, after cleaning, are allowed
	long := strings.Repeat("я", 200)
	if _, err := service.PatchIngredient(&PatchIngredientRequest{IngredientID: id, UserID: user, Name: ptr(" " + long + " ")}); err != nil {
		t.Fatalf("PatchIngredient with a 200-character name: %v", err)
	}
	assertIngredient(t, repo, id, long, 250, nil, nil, nil)
}

func TestPatchIngredientNothingToChange(t *testing.T) {
	service, repo := newTestService()
	id := repo.insert(user, "Хлеб", 250, nil, nil, nil)

	_, err := service.PatchIngredient(&PatchIngredientRequest{IngredientID: id, UserID: user})
	assertValidationError(t, err, "nothing to change", "name", "kcal_per_100g", "fats", "carbs", "proteins")
}

func TestPatchIngredientKeepsStoredValuesOutsideLimits(t *testing.T) {
	service, repo := newTestService()
	// The website accepts any value, a rename must not be blocked by it
	id := repo.insert(user, "Масло", 1000, ptr(120.0), nil, nil)

	if _, err := service.PatchIngredient(&PatchIngredientRequest{IngredientID: id, UserID: user, Name: ptr("Масло сливочное")}); err != nil {
		t.Fatalf("PatchIngredient: %v", err)
	}
	assertIngredient(t, repo, id, "Масло сливочное", 1000, 120.0, nil, nil)
}

func TestPatchIngredientNotFound(t *testing.T) {
	service, repo := newTestService()
	foreignID := repo.insert(otherUser, "Хлеб", 250, nil, nil, nil)

	for _, id := range []int{foreignID, 999} {
		_, err := service.PatchIngredient(&PatchIngredientRequest{IngredientID: id, UserID: user, KcalPer100g: ptr(1.0)})
		if !errors.Is(err, sql.ErrNoRows) {
			t.Errorf("ingredient %d: error = %v, want sql.ErrNoRows", id, err)
		}
	}
	_, err := service.UpdateIngredient(&UpdateIngredientRequest{IngredientID: foreignID, UserID: user, Name: "Хлеб", KcalPer100g: 1})
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("full update of a foreign ingredient: error = %v, want sql.ErrNoRows", err)
	}
	assertIngredient(t, repo, foreignID, "Хлеб", 250, nil, nil, nil)
}

// nameVariants differ from "Гречка" only by case, spacing or invisible characters
var nameVariants = []string{"Гречка", "гречка", " Гречка ", "ГРЕЧКА", "Греч" + zeroWidthSpace + "ка"}

func TestCreateIngredientRejectsDuplicateNames(t *testing.T) {
	service, repo := newTestService()
	buckwheatID := repo.insert(user, "Гречка", 330, nil, nil, nil)
	hedgehogID := repo.insert(user, "Ёжик", 100, nil, nil, nil)
	repo.insert(otherUser, "Овсянка", 370, nil, nil, nil)

	for _, name := range nameVariants {
		_, err := service.CreateIngredient(&CreateIngredientRequest{UserID: user, Name: name, KcalPer100g: 330})
		assertDuplicate(t, err, buckwheatID)
	}
	_, err := service.CreateIngredient(&CreateIngredientRequest{UserID: user, Name: "ежик", KcalPer100g: 100})
	assertDuplicate(t, err, hedgehogID)
	if len(repo.ingredients) != 3 {
		t.Errorf("ingredients = %d, want no new ones", len(repo.ingredients))
	}

	// Names of another user do not conflict, the name is stored as sent
	created, err := service.CreateIngredient(&CreateIngredientRequest{UserID: user, Name: "Овсянка", KcalPer100g: 370})
	if err != nil || created.Name != "Овсянка" || created.UserID != user {
		t.Errorf("CreateIngredient = %+v, %v; want a new ingredient of the user", created, err)
	}
}

func TestUpdateIngredientRejectsDuplicateNames(t *testing.T) {
	update := func(service *Service, id int, name string) error {
		_, err := service.UpdateIngredient(&UpdateIngredientRequest{IngredientID: id, UserID: user, Name: name, KcalPer100g: 260})
		return err
	}
	patch := func(service *Service, id int, name string) error {
		_, err := service.PatchIngredient(&PatchIngredientRequest{IngredientID: id, UserID: user, Name: ptr(name), KcalPer100g: ptr(260.0)})
		return err
	}

	for method, change := range map[string]func(*Service, int, string) error{"full update": update, "partial update": patch} {
		t.Run(method, func(t *testing.T) {
			service, repo := newTestService()
			buckwheatID := repo.insert(user, "Гречка", 330, nil, nil, nil)
			hedgehogID := repo.insert(user, "Ёжик", 100, nil, nil, nil)
			breadID := repo.insert(user, "Хлеб", 250, ptr(1.0), nil, nil)
			repo.insert(otherUser, "Овсянка", 370, nil, nil, nil)

			for _, name := range nameVariants {
				assertDuplicate(t, change(service, breadID, name), buckwheatID)
			}
			assertDuplicate(t, change(service, breadID, "ежик"), hedgehogID)
			assertIngredient(t, repo, breadID, "Хлеб", 250, 1.0, nil, nil)

			// A variant of the own name is not a conflict
			for _, name := range []string{"хлеб", "ХЛЕБ"} {
				if err := change(service, breadID, name); err != nil {
					t.Errorf("rename to %q: %v", name, err)
				}
			}
			if err := change(service, hedgehogID, "ежик"); err != nil {
				t.Errorf("rename Ёжик to ежик: %v", err)
			}

			// A name of another user is not a conflict
			if err := change(service, breadID, "Овсянка"); err != nil {
				t.Errorf("rename to another user's name: %v", err)
			}
			if repo.ingredients[breadID].Name != "Овсянка" {
				t.Errorf("name = %q, want Овсянка", repo.ingredients[breadID].Name)
			}
		})
	}
}

func TestUpdateIngredientWithPreExistingNearDuplicate(t *testing.T) {
	service, repo := newTestService()
	// Both were created before the duplicate rule existed
	riceID := repo.insert(user, "Рис", 130, nil, nil, nil)
	lowerRiceID := repo.insert(user, "рис", 130, nil, nil, nil)
	breadID := repo.insert(user, "Хлеб", 250, nil, nil, nil)

	// Editing without changing the normalized name is allowed
	if _, err := service.UpdateIngredient(&UpdateIngredientRequest{IngredientID: lowerRiceID, UserID: user, Name: "рис", KcalPer100g: 120}); err != nil {
		t.Fatalf("full update: %v", err)
	}
	if _, err := service.PatchIngredient(&PatchIngredientRequest{IngredientID: lowerRiceID, UserID: user, KcalPer100g: ptr(125.0)}); err != nil {
		t.Fatalf("partial update: %v", err)
	}
	if _, err := service.PatchIngredient(&PatchIngredientRequest{IngredientID: lowerRiceID, UserID: user, Name: ptr("РИС")}); err != nil {
		t.Fatalf("case-only rename: %v", err)
	}
	assertIngredient(t, repo, lowerRiceID, "РИС", 125, nil, nil, nil)

	// The exact name of the other one is rejected by the unique index and reported as a duplicate, not as a failure
	_, err := service.PatchIngredient(&PatchIngredientRequest{IngredientID: lowerRiceID, UserID: user, Name: ptr("Рис")})
	assertDuplicate(t, err, riceID)
	assertIngredient(t, repo, lowerRiceID, "РИС", 125, nil, nil, nil)

	// Another ingredient cannot take that name
	_, err = service.UpdateIngredient(&UpdateIngredientRequest{IngredientID: breadID, UserID: user, Name: "рис", KcalPer100g: 250})
	assertDuplicate(t, err, riceID)
	assertIngredient(t, repo, breadID, "Хлеб", 250, nil, nil, nil)
}

func TestDuplicateCreatedConcurrently(t *testing.T) {
	service, repo := newTestService()
	breadID := repo.insert(user, "Хлеб", 250, nil, nil, nil)

	// Another request adds the same name between the check and the write: the unique index catches it
	var concurrentID int
	repo.beforeWrite = func() {
		concurrentID = repo.insert(user, "Гречка", 330, nil, nil, nil)
		repo.beforeWrite = nil
	}
	_, err := service.CreateIngredient(&CreateIngredientRequest{UserID: user, Name: "Гречка", KcalPer100g: 330})
	assertDuplicate(t, err, concurrentID)

	repo.beforeWrite = func() {
		concurrentID = repo.insert(user, "Рис", 130, nil, nil, nil)
		repo.beforeWrite = nil
	}
	_, err = service.PatchIngredient(&PatchIngredientRequest{IngredientID: breadID, UserID: user, Name: ptr("Рис")})
	assertDuplicate(t, err, concurrentID)
	assertIngredient(t, repo, breadID, "Хлеб", 250, nil, nil, nil)
}

func TestWriteErrorsAreNotDuplicates(t *testing.T) {
	service, repo := newTestService()
	breadID := repo.insert(user, "Хлеб", 250, nil, nil, nil)
	dbErr := errors.New("database is locked")

	repo.writeErr = dbErr
	if _, err := service.CreateIngredient(&CreateIngredientRequest{UserID: user, Name: "Гречка", KcalPer100g: 330}); !errors.Is(err, dbErr) {
		t.Errorf("create: error = %v, want the database error", err)
	}
	repo.writeErr = dbErr
	if _, err := service.PatchIngredient(&PatchIngredientRequest{IngredientID: breadID, UserID: user, Name: ptr("Батон")}); !errors.Is(err, dbErr) {
		t.Errorf("patch: error = %v, want the database error", err)
	}

	// A near-duplicate from before the rule cannot make a write fail (the unique index compares the exact name),
	// so a failed calories-only edit is the database error, not a name conflict
	repo.insert(user, "Рис", 130, nil, nil, nil)
	lowerRiceID := repo.insert(user, "рис", 130, nil, nil, nil)
	repo.writeErr = dbErr
	if _, err := service.PatchIngredient(&PatchIngredientRequest{IngredientID: lowerRiceID, UserID: user, KcalPer100g: ptr(120.0)}); !errors.Is(err, dbErr) {
		t.Errorf("patch of a near-duplicate: error = %v, want the database error", err)
	}
	repo.writeErr = dbErr
	if _, err := service.UpdateIngredient(&UpdateIngredientRequest{IngredientID: lowerRiceID, UserID: user, Name: "рис", KcalPer100g: 120}); !errors.Is(err, dbErr) {
		t.Errorf("full update of a near-duplicate: error = %v, want the database error", err)
	}
	assertIngredient(t, repo, lowerRiceID, "рис", 130, nil, nil, nil)

	// The duplicate check cannot run: the write does not happen
	repo.listErr = dbErr
	if _, err := service.CreateIngredient(&CreateIngredientRequest{UserID: user, Name: "Гречка", KcalPer100g: 330}); !errors.Is(err, dbErr) {
		t.Errorf("create without the list: error = %v, want the database error", err)
	}
	if len(repo.ingredients) != 3 {
		t.Errorf("ingredients = %d, want only the original ones", len(repo.ingredients))
	}
}

// addRequest is a valid new ingredient of the external API; change modifies it before it is sent
func addRequest(change func(*AddIngredientRequest)) *AddIngredientRequest {
	req := &AddIngredientRequest{UserID: user, Name: ptr("Хлебцы ржаные"), KcalPer100g: ptr(310.0),
		Fats: ptr(2.0), Carbs: ptr(60.0), Proteins: ptr(10.0)}
	if change != nil {
		change(req)
	}
	return req
}

func TestAddIngredientCreatesIngredient(t *testing.T) {
	service, repo := newTestService()

	got, err := service.AddIngredient(addRequest(nil))
	if err != nil {
		t.Fatalf("AddIngredient: %v", err)
	}
	if got.UserID != user || got.Name != "Хлебцы ржаные" || got.TimesUsed != 0 || got.LastUsed != nil {
		t.Errorf("returned ingredient = %+v", got)
	}
	assertIngredient(t, repo, got.ID, "Хлебцы ржаные", 310, 2.0, 60.0, 10.0)

	// Nutrients that are not sent stay unknown; the cleaned name is stored
	got, err = service.AddIngredient(&AddIngredientRequest{
		UserID: user, Name: ptr("  Хлебцы" + zeroWidthSpace + "   с отрубями "), KcalPer100g: ptr(120.0), Proteins: ptr(0.0),
	})
	if err != nil {
		t.Fatalf("AddIngredient without fats and carbs: %v", err)
	}
	if got.Name != "Хлебцы с отрубями" {
		t.Errorf("name = %q, want the cleaned name", got.Name)
	}
	assertIngredient(t, repo, got.ID, "Хлебцы с отрубями", 120, nil, nil, 0.0)
}

func TestAddIngredientLimits(t *testing.T) {
	service, repo := newTestService()

	longName := strings.Repeat("я", 200)
	valid := []struct {
		name                  string
		kcal                  float64
		fats, carbs, proteins *float64
	}{
		{"Вода", 0, nil, nil, nil},
		{"Масло", 900, nil, nil, nil},
		{"Ноль", 100, ptr(0.0), ptr(0.0), ptr(0.0)},
		{"Сто", 100, ptr(100.0), ptr(100.0), ptr(100.0)},
		// 200 characters after cleaning
		{" " + longName + " ", 100, nil, nil, nil},
	}
	for _, v := range valid {
		got, err := service.AddIngredient(&AddIngredientRequest{
			UserID: user, Name: ptr(v.name), KcalPer100g: ptr(v.kcal), Fats: v.fats, Carbs: v.carbs, Proteins: v.proteins,
		})
		if err != nil {
			t.Errorf("%q, %g kcal: %v", v.name, v.kcal, err)
			continue
		}
		assertIngredient(t, repo, got.ID, strings.TrimSpace(v.name), v.kcal, floatValue(v.fats), floatValue(v.carbs), floatValue(v.proteins))
	}
	created := len(repo.ingredients)

	invalid := []struct {
		change    func(*AddIngredientRequest)
		wantParts []string
	}{
		{func(r *AddIngredientRequest) { r.KcalPer100g = ptr(-1.0) }, []string{"kcal_per_100g must be between 0 and 900"}},
		{func(r *AddIngredientRequest) { r.KcalPer100g = ptr(-0.1) }, []string{"kcal_per_100g must be between 0 and 900"}},
		{func(r *AddIngredientRequest) { r.KcalPer100g = ptr(900.1) }, []string{"kcal_per_100g must be between 0 and 900"}},
		{func(r *AddIngredientRequest) { r.KcalPer100g = nil }, []string{"kcal_per_100g is required"}},
		{func(r *AddIngredientRequest) { r.Fats = ptr(-0.1) }, []string{"fats must be between 0 and 100 (grams per 100 g)"}},
		{func(r *AddIngredientRequest) { r.Fats = ptr(100.1) }, []string{"fats must be between 0 and 100 (grams per 100 g)"}},
		{func(r *AddIngredientRequest) { r.Carbs = ptr(-0.1) }, []string{"carbs must be between 0 and 100 (grams per 100 g)"}},
		{func(r *AddIngredientRequest) { r.Carbs = ptr(100.1) }, []string{"carbs must be between 0 and 100 (grams per 100 g)"}},
		{func(r *AddIngredientRequest) { r.Proteins = ptr(-0.1) }, []string{"proteins must be between 0 and 100 (grams per 100 g)"}},
		{func(r *AddIngredientRequest) { r.Proteins = ptr(100.1) }, []string{"proteins must be between 0 and 100 (grams per 100 g)"}},
		{func(r *AddIngredientRequest) { r.Name = nil }, []string{"name is required"}},
		{func(r *AddIngredientRequest) { r.Name = ptr("") }, []string{"name is required"}},
		{func(r *AddIngredientRequest) { r.Name = ptr("   ") }, []string{"name is required"}},
		{func(r *AddIngredientRequest) { r.Name = ptr(zeroWidthSpace + " \t" + zeroWidthSpace) }, []string{"name is required"}},
		{func(r *AddIngredientRequest) { r.Name = ptr(strings.Repeat("я", 201)) }, []string{"name must be at most 200 characters"}},
	}
	for _, tt := range invalid {
		_, err := service.AddIngredient(addRequest(tt.change))
		assertValidationError(t, err, tt.wantParts...)
	}
	if len(repo.ingredients) != created {
		t.Errorf("ingredients = %d, want %d: a rejected ingredient was created", len(repo.ingredients), created)
	}
}

func TestAddIngredientRejectsDuplicateNames(t *testing.T) {
	service, repo := newTestService()
	buckwheatID := repo.insert(user, "Гречка", 330, nil, nil, nil)
	hedgehogID := repo.insert(user, "Ёжик", 100, nil, nil, nil)
	repo.insert(otherUser, "Овсянка", 370, nil, nil, nil)

	for _, name := range nameVariants {
		_, err := service.AddIngredient(addRequest(func(r *AddIngredientRequest) { r.Name = ptr(name) }))
		assertDuplicate(t, err, buckwheatID)
	}
	_, err := service.AddIngredient(addRequest(func(r *AddIngredientRequest) { r.Name = ptr("ежик") }))
	assertDuplicate(t, err, hedgehogID)
	if len(repo.ingredients) != 3 {
		t.Errorf("ingredients = %d, want no new ones", len(repo.ingredients))
	}
	assertIngredient(t, repo, buckwheatID, "Гречка", 330, nil, nil, nil)

	// Names of another user do not conflict
	created, err := service.AddIngredient(addRequest(func(r *AddIngredientRequest) { r.Name = ptr("Овсянка") }))
	if err != nil || created.Name != "Овсянка" || created.UserID != user {
		t.Errorf("AddIngredient = %+v, %v; want a new ingredient of the user", created, err)
	}
}

func TestAddIngredientUsage(t *testing.T) {
	service, repo := newTestService()
	// One-off diary entries with exactly this name already exist: the list counts them for the new ingredient
	repo.diary = map[string]models.UserIngredientUsage{"Кофе раф": {TimesUsed: 2, LastUsed: ptr("2026-10-15")}}

	got, err := service.AddIngredient(addRequest(func(r *AddIngredientRequest) { r.Name = ptr(" Кофе раф ") }))
	if err != nil {
		t.Fatalf("AddIngredient: %v", err)
	}
	if got.Name != "Кофе раф" || got.TimesUsed != 2 || got.LastUsed == nil || *got.LastUsed != "2026-10-15" {
		t.Errorf("returned ingredient = %+v, want the usage of the diary entries", got)
	}

	// The ingredient is stored; a failure to read the statistics does not turn the create into an error
	repo.usageErr = errors.New("database is locked")
	got, err = service.AddIngredient(addRequest(nil))
	if err != nil {
		t.Fatalf("AddIngredient without usage: %v", err)
	}
	if got.ID == 0 || got.Name != "Хлебцы ржаные" || got.TimesUsed != 0 || got.LastUsed != nil {
		t.Errorf("returned ingredient = %+v, want it without usage", got)
	}
	assertIngredient(t, repo, got.ID, "Хлебцы ржаные", 310, 2.0, 60.0, 10.0)

	// Deleted by a concurrent request between the create and the statistics: the created ingredient is returned
	repo.usageErr = nil
	repo.beforeUsage = func() { delete(repo.ingredients, repo.nextID-1) }
	got, err = service.AddIngredient(addRequest(func(r *AddIngredientRequest) { r.Name = ptr("Батон") }))
	if err != nil {
		t.Fatalf("AddIngredient deleted concurrently: %v", err)
	}
	if got.ID != repo.nextID-1 || got.Name != "Батон" || got.KcalPer100g != 310 || got.TimesUsed != 0 || got.LastUsed != nil {
		t.Errorf("returned ingredient = %+v, want the created one without usage", got)
	}
}

func TestAddIngredientWriteErrors(t *testing.T) {
	service, repo := newTestService()
	dbErr := errors.New("database is locked")

	// Another request adds exactly the same name between the check and the write: the unique index catches it
	var concurrentID int
	repo.beforeWrite = func() {
		concurrentID = repo.insert(user, "Хлебцы ржаные", 300, nil, nil, nil)
		repo.beforeWrite = nil
	}
	_, err := service.AddIngredient(addRequest(nil))
	assertDuplicate(t, err, concurrentID)

	// A plain write failure is not a duplicate
	repo.writeErr = dbErr
	if _, err := service.AddIngredient(addRequest(func(r *AddIngredientRequest) { r.Name = ptr("Гречка") })); !errors.Is(err, dbErr) {
		t.Errorf("error = %v, want the database error", err)
	}

	// A near-duplicate added concurrently cannot make the write fail (the index compares the exact name),
	// so a failed write is the database error, not a name conflict
	repo.beforeWrite = func() {
		repo.insert(user, "рис", 130, nil, nil, nil)
		repo.beforeWrite = nil
	}
	repo.writeErr = dbErr
	_, err = service.AddIngredient(addRequest(func(r *AddIngredientRequest) { r.Name = ptr("Рис") }))
	var duplicateErr *DuplicateIngredientError
	if !errors.Is(err, dbErr) || errors.As(err, &duplicateErr) {
		t.Errorf("near-duplicate: error = %v, want the database error", err)
	}

	// The duplicate check cannot run: the write does not happen
	repo.listErr = dbErr
	if _, err := service.AddIngredient(addRequest(func(r *AddIngredientRequest) { r.Name = ptr("Батон") })); !errors.Is(err, dbErr) {
		t.Errorf("without the list: error = %v, want the database error", err)
	}
	if len(repo.ingredients) != 2 {
		t.Errorf("ingredients = %d, want only the concurrent ones", len(repo.ingredients))
	}
}

// The website create keeps its rules: no upper limits and the name stored as sent
func TestCreateIngredientKeepsWebsiteRules(t *testing.T) {
	service, repo := newTestService()

	created, err := service.CreateIngredient(&CreateIngredientRequest{UserID: user, Name: " Масло ", KcalPer100g: 1000, Fats: ptr(120.0)})
	if err != nil {
		t.Fatalf("CreateIngredient: %v", err)
	}
	assertIngredient(t, repo, created.ID, " Масло ", 1000, 120.0, nil, nil)

	for _, req := range []CreateIngredientRequest{
		{UserID: user, Name: "", KcalPer100g: 100},
		{UserID: user, Name: "Хлеб", KcalPer100g: -1},
		{UserID: user, Name: "Хлеб", KcalPer100g: 100, Carbs: ptr(-1.0)},
	} {
		if _, err := service.CreateIngredient(&req); err == nil {
			t.Errorf("%+v: want a validation error", req)
		}
	}
	if len(repo.ingredients) != 1 {
		t.Errorf("ingredients = %d, want 1", len(repo.ingredients))
	}
}
