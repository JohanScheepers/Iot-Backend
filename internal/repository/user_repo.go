package repository

import (
	"context"
	"database/sql"
)

type UserRepo struct {
	DB *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{DB: db}
}

func (r *UserRepo) GetByFirebaseUID(ctx context.Context, uid string) (*User, error) {
	var u User
	query := "SELECT id, firebase_uid, email, company_id, created_at FROM users WHERE firebase_uid = ?"
	err := r.DB.QueryRowContext(ctx, query, uid).Scan(&u.ID, &u.FirebaseUID, &u.Email, &u.CompanyID, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
