package model

// User represents the users table in PostgreSQL
type User struct {
	Id        int64  `db:"id"`
	Username  string `db:"username"`
	Password  string `db:"password"` // Hashed password
	Salt      string `db:"salt"`     // Password salt
	Email     string `db:"email"`
	Phone     string `db:"phone"`
	Avatar    string `db:"avatar"`
	Role      string `db:"role"`       // user, admin
	Status    int64  `db:"status"`     // 1:active, 2:disabled
	CreatedAt int64  `db:"created_at"` // Unix timestamp
	UpdatedAt int64  `db:"updated_at"`
}
