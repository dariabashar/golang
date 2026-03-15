package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
	GenderOther  Gender = "other"
)

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Gender    Gender    `json:"gender"`
	BirthDate time.Time `json:"birthDate"`
}

type PaginatedResponse struct {
	Data       []User `json:"data"`
	TotalCount int    `json:"totalCount"`
	Page       int    `json:"page"`
	PageSize   int    `json:"pageSize"`
}

type UserFilter struct {
	ID        *int
	Name      *string
	Email     *string
	Gender    *Gender
	BirthDate *time.Time
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetPaginatedUsers returns paginated users with dynamic filtering and ordering.
// All logic is done in a single SQL query using WHERE/ORDER BY/LIMIT/OFFSET.
func (r *Repository) GetPaginatedUsers(ctx context.Context, page, pageSize int, filter UserFilter, orderBy string, orderDesc bool) (PaginatedResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	clauses := make([]string, 0)
	args := make([]any, 0)

	// dynamic filters
	if filter.ID != nil {
		clauses = append(clauses, fmt.Sprintf("u.id = $%d", len(args)+1))
		args = append(args, *filter.ID)
	}
	if filter.Name != nil {
		clauses = append(clauses, fmt.Sprintf("u.name ILIKE $%d", len(args)+1))
		args = append(args, "%"+*filter.Name+"%")
	}
	if filter.Email != nil {
		clauses = append(clauses, fmt.Sprintf("u.email ILIKE $%d", len(args)+1))
		args = append(args, "%"+*filter.Email+"%")
	}
	if filter.Gender != nil {
		clauses = append(clauses, fmt.Sprintf("u.gender = $%d", len(args)+1))
		args = append(args, *filter.Gender)
	}
	if filter.BirthDate != nil {
		clauses = append(clauses, fmt.Sprintf("u.birth_date = $%d", len(args)+1))
		args = append(args, *filter.BirthDate)
	}

	where := ""
	if len(clauses) > 0 {
		where = "WHERE " + strings.Join(clauses, " AND ")
	}

	// order_by with whitelisting
	allowedOrder := map[string]string{
		"id":         "u.id",
		"name":       "u.name",
		"email":      "u.email",
		"gender":     "u.gender",
		"birth_date": "u.birth_date",
	}
	col, ok := allowedOrder[strings.ToLower(orderBy)]
	if !ok {
		col = "u.id"
	}
	dir := "ASC"
	if orderDesc {
		dir = "DESC"
	}

	// single query returns rows plus total_count using window function
	query := fmt.Sprintf(`
		SELECT
			u.id,
			u.name,
			u.email,
			u.gender,
			u.birth_date,
			COUNT(*) OVER() AS total_count
		FROM users u
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, where, col, dir, len(args)+1, len(args)+2)

	args = append(args, pageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return PaginatedResponse{}, err
	}
	defer rows.Close()

	var users []User
	totalCount := 0

	for rows.Next() {
		var u User
		var count int
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Gender, &u.BirthDate, &count); err != nil {
			return PaginatedResponse{}, err
		}
		users = append(users, u)
		totalCount = count
	}
	if rows.Err() != nil {
		return PaginatedResponse{}, rows.Err()
	}

	return PaginatedResponse{
		Data:       users,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

// GetCommonFriends returns common friends of two users avoiding N+1 problem.
func (r *Repository) GetCommonFriends(ctx context.Context, userID1, userID2 int) ([]User, error) {
	// One SQL query with JOINs to get mutual friends
	const q = `
		SELECT u.id, u.name, u.email, u.gender, u.birth_date
		FROM users u
		JOIN user_friends uf1 ON uf1.friend_id = u.id
		JOIN user_friends uf2 ON uf2.friend_id = u.id
		WHERE uf1.user_id = $1
		  AND uf2.user_id = $2
		  AND uf1.user_id <> uf2.user_id
		ORDER BY u.id
	`

	rows, err := r.db.QueryContext(ctx, q, userID1, userID2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Gender, &u.BirthDate); err != nil {
			return nil, err
		}
		res = append(res, u)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return res, nil
}

