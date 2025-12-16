package model

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserModel = (*customUserModel)(nil)

type (
	// UserModel is an interface to be customized
	UserModel interface {
		// Insert a new user
		Insert(data *User) (sql.Result, error)

		// FindOne by user ID
		FindOne(id int64) (*User, error)

		// FindOneByUsername finds user by username
		FindOneByUsername(username string) (*User, error)

		// FindOneByEmail finds user by email
		FindOneByEmail(email string) (*User, error)

		// Update user profile
		Update(data *User) error

		// Delete user (soft delete by setting status)
		Delete(id int64) error
	}

	customUserModel struct {
		conn sqlx.SqlConn
	}
)

// NewUserModel returns a UserModel instance
func NewUserModel(conn sqlx.SqlConn) UserModel {
	return &customUserModel{
		conn: conn,
	}
}

// Insert inserts a new user into database
func (m *customUserModel) Insert(data *User) (sql.Result, error) {
	query := `INSERT INTO users (username, password, salt, email, phone, avatar, status, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			  RETURNING id`

	var id int64
	err := m.conn.QueryRow(&id, query,
		data.Username,
		data.Password,
		data.Salt,
		data.Email,
		data.Phone,
		data.Avatar,
		data.Status,
		data.CreatedAt,
		data.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	data.Id = id
	return &insertResult{lastInsertId: id}, nil
}

// FindOne finds user by ID
func (m *customUserModel) FindOne(id int64) (*User, error) {
	query := `SELECT id, username, password, salt, email, phone, avatar, status, created_at, updated_at
			  FROM users
			  WHERE id = $1 AND status = 1`

	var user User
	err := m.conn.QueryRow(&user, query, id)

	switch err {
	case nil:
		return &user, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// FindOneByUsername finds user by username
func (m *customUserModel) FindOneByUsername(username string) (*User, error) {
	query := `SELECT id, username, password, salt, email, phone, avatar, status, created_at, updated_at
			  FROM users
			  WHERE username = $1`

	var user User
	err := m.conn.QueryRow(&user, query, username)

	switch err {
	case nil:
		return &user, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// FindOneByEmail finds user by email
func (m *customUserModel) FindOneByEmail(email string) (*User, error) {
	query := `SELECT id, username, password, salt, email, phone, avatar, status, created_at, updated_at
			  FROM users
			  WHERE email = $1`

	var user User
	err := m.conn.QueryRow(&user, query, email)

	switch err {
	case nil:
		return &user, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// Update updates user profile
func (m *customUserModel) Update(data *User) error {
	query := `UPDATE users
			  SET email = $1, phone = $2, avatar = $3, updated_at = $4
			  WHERE id = $5`

	_, err := m.conn.Exec(query, data.Email, data.Phone, data.Avatar, data.UpdatedAt, data.Id)
	return err
}

// Delete soft deletes user by setting status to 2
func (m *customUserModel) Delete(id int64) error {
	query := `UPDATE users SET status = 2 WHERE id = $1`
	_, err := m.conn.Exec(query, id)
	return err
}

// ErrNotFound is returned when a user is not found
var ErrNotFound = sqlx.ErrNotFound

// insertResult implements sql.Result for Insert operation
type insertResult struct {
	lastInsertId int64
}

func (r *insertResult) LastInsertId() (int64, error) {
	return r.lastInsertId, nil
}

func (r *insertResult) RowsAffected() (int64, error) {
	return 1, nil
}

// BuildUpdateQuery dynamically builds UPDATE query for partial updates
func BuildUpdateQuery(tableName string, id int64, updates map[string]interface{}) (string, []interface{}) {
	if len(updates) == 0 {
		return "", nil
	}

	var setClauses []string
	var values []interface{}
	paramCount := 1

	for key, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", key, paramCount))
		values = append(values, value)
		paramCount++
	}

	values = append(values, id)
	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = $%d",
		tableName,
		strings.Join(setClauses, ", "),
		paramCount,
	)

	return query, values
}
