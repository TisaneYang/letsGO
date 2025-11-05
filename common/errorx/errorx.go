package errorx

import "fmt"

// ========================================
// Custom Error Type for Business Logic
// ========================================
// This allows us to attach error codes to errors
// and handle them appropriately in the API layer

type CodeError struct {
	Code int    // Error code
	Msg  string // Error message
}

// Error implements the error interface
func (e *CodeError) Error() string {
	return fmt.Sprintf("code: %d, msg: %s", e.Code, e.Msg)
}

// GetCode returns the error code
func (e *CodeError) GetCode() int {
	return e.Code
}

// GetMsg returns the error message
func (e *CodeError) GetMsg() string {
	return e.Msg
}

// NewCodeError creates a new CodeError
func NewCodeError(code int, msg string) *CodeError {
	return &CodeError{
		Code: code,
		Msg:  msg,
	}
}

// ========================================
// Default Errors
// ========================================

var (
	ErrSystem         = NewCodeError(1000, "System error")
	ErrInvalidParams  = NewCodeError(1001, "Invalid parameters")
	ErrDatabase       = NewCodeError(1002, "Database error")
	ErrCache          = NewCodeError(1003, "Cache error")
	ErrRPC            = NewCodeError(1004, "RPC call error")

	ErrUserNotFound      = NewCodeError(2000, "User not found")
	ErrUserExists        = NewCodeError(2001, "User already exists")
	ErrWrongPassword     = NewCodeError(2002, "Wrong password")
	ErrTokenInvalid      = NewCodeError(2003, "Invalid token")
	ErrTokenExpired      = NewCodeError(2004, "Token expired")
	ErrPermissionDenied  = NewCodeError(2005, "Permission denied")

	ErrProductNotFound   = NewCodeError(3000, "Product not found")
	ErrProductOutOfStock = NewCodeError(3001, "Product out of stock")

	ErrCartEmpty         = NewCodeError(4000, "Cart is empty")
	ErrCartItemNotFound  = NewCodeError(4001, "Cart item not found")

	ErrOrderNotFound     = NewCodeError(5000, "Order not found")
	ErrOrderCannotCancel = NewCodeError(5002, "Cannot cancel order")

	ErrPaymentNotFound   = NewCodeError(6000, "Payment not found")
	ErrPaymentFailed     = NewCodeError(6001, "Payment failed")
)
