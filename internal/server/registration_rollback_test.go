package server

import (
	"net"
	"net/http"
	"testing"

	"ypeskov/kkal-tracker/internal/config"
)

// closedPort returns a local TCP port nothing listens on
func closedPort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to reserve a port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	return port
}

// A registration whose activation email fails is rolled back: deleting the user cascades to the activation token
// and to the copied global ingredients
func TestRegistrationRollbackLeavesNoRows(t *testing.T) {
	smtpPort := closedPort(t)
	client, _ := newSmokeEnvironmentWith(t, discardLogger(), func(cfg *config.Config) {
		cfg.SMTPPort = smtpPort
	})

	// A global ingredient, so the registration copies one into user_ingredients
	result, err := client.db.Exec(`INSERT INTO global_ingredients (kcal_per_100g) VALUES (52)`)
	if err != nil {
		t.Fatalf("failed to insert a global ingredient: %v", err)
	}
	ingredientID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("failed to read the global ingredient id: %v", err)
	}
	if _, err := client.db.Exec(`INSERT INTO global_ingredient_names (ingredient_id, language_code, name)
		VALUES (?, 'en_US', 'Apple')`, ingredientID); err != nil {
		t.Fatalf("failed to insert a global ingredient name: %v", err)
	}

	const email = "rollback@example.com"
	client.expect(http.StatusServiceUnavailable, http.MethodPost, "/api/auth/register",
		map[string]string{"email": email, "password": "smoke-password", "language_code": "en_US"}, nil)

	count := func(query string, args ...any) int {
		t.Helper()
		var n int
		if err := client.db.QueryRow(query, args...).Scan(&n); err != nil {
			t.Fatalf("%s: %v", query, err)
		}
		return n
	}

	// The AUTOINCREMENT counters prove the user, the token and the ingredient copy were written before the rollback
	for _, table := range []string{"users", "activation_tokens", "user_ingredients"} {
		if n := count(`SELECT COUNT(*) FROM sqlite_sequence WHERE name = ? AND seq > 0`, table); n != 1 {
			t.Errorf("%s: no row was ever inserted, the rollback is not exercised", table)
		}
	}

	if n := count(`SELECT COUNT(*) FROM users WHERE email = ?`, email); n != 0 {
		t.Errorf("users: %d rows remain for %s", n, email)
	}
	// The only user was the rolled back one, so every remaining row would belong to it
	for _, table := range []string{"activation_tokens", "user_ingredients"} {
		if n := count(`SELECT COUNT(*) FROM ` + table); n != 0 {
			t.Errorf("%s: %d rows of the rolled back user remain", table, n)
		}
	}
}
