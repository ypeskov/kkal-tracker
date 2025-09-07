package models

import (
	"database/sql"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(email, password string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO users (email, password_hash)
		VALUES (?, ?)
	`
	
	result, err := r.db.Exec(query, email, string(hashedPassword))
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.GetByID(int(id))
}

func (r *UserRepository) GetByID(id int) (*User, error) {
	query := `
		SELECT id, email, password_hash, created_at, updated_at
		FROM users
		WHERE id = ?
	`
	
	user := &User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return user, nil
}

func (r *UserRepository) GetByEmail(email string) (*User, error) {
	query := `
		SELECT id, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = ?
	`
	
	user := &User{}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return user, nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}