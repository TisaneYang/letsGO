package response

// ========================================
// Standard API Response Structure
// ========================================
// All API responses follow this unified format
// This ensures consistency across all endpoints

// Success Response
// Example: {"code": 0, "msg": "success", "data": {...}}
type Response struct {
	Code int         `json:"code"` // 0 = success, non-zero = error
	Msg  string      `json:"msg"`  // Message description
	Data interface{} `json:"data"` // Response data (can be any type)
}

// List Response (for paginated data)
// Example: {"code": 0, "msg": "success", "data": {"list": [...], "total": 100, "page": 1, "pageSize": 20}}
type ListResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// ========================================
// Response Helper Functions
// ========================================

// Success returns a successful response
func Success(data interface{}) *Response {
	return &Response{
		Code: 0,
		Msg:  "success",
		Data: data,
	}
}

// Error returns an error response with custom message
func Error(code int, msg string) *Response {
	return &Response{
		Code: code,
		Msg:  msg,
		Data: nil,
	}
}

// ========================================
// Common Error Codes
// ========================================
const (
	// Success
	SUCCESS = 0

	// System errors (1000-1999)
	ERROR_SYSTEM           = 1000 // System error
	ERROR_INVALID_PARAMS   = 1001 // Invalid parameters
	ERROR_DATABASE         = 1002 // Database error
	ERROR_CACHE            = 1003 // Cache error
	ERROR_RPC              = 1004 // RPC call error
	ERROR_THIRD_PARTY      = 1005 // Third-party service error

	// User errors (2000-2999)
	ERROR_USER_NOT_FOUND       = 2000 // User not found
	ERROR_USER_ALREADY_EXISTS  = 2001 // User already exists
	ERROR_WRONG_PASSWORD       = 2002 // Wrong password
	ERROR_TOKEN_INVALID        = 2003 // Invalid token
	ERROR_TOKEN_EXPIRED        = 2004 // Token expired
	ERROR_PERMISSION_DENIED    = 2005 // Permission denied

	// Product errors (3000-3999)
	ERROR_PRODUCT_NOT_FOUND    = 3000 // Product not found
	ERROR_PRODUCT_OUT_OF_STOCK = 3001 // Product out of stock
	ERROR_PRODUCT_INVALID      = 3002 // Invalid product

	// Cart errors (4000-4999)
	ERROR_CART_EMPTY           = 4000 // Cart is empty
	ERROR_CART_ITEM_NOT_FOUND  = 4001 // Cart item not found
	ERROR_CART_EXCEED_LIMIT    = 4002 // Cart exceeds limit

	// Order errors (5000-5999)
	ERROR_ORDER_NOT_FOUND      = 5000 // Order not found
	ERROR_ORDER_STATUS_INVALID = 5001 // Invalid order status
	ERROR_ORDER_CANNOT_CANCEL  = 5002 // Cannot cancel order
	ERROR_ORDER_CREATE_FAILED  = 5003 // Order creation failed

	// Payment errors (6000-6999)
	ERROR_PAYMENT_NOT_FOUND    = 6000 // Payment not found
	ERROR_PAYMENT_FAILED       = 6001 // Payment failed
	ERROR_PAYMENT_TIMEOUT      = 6002 // Payment timeout
	ERROR_PAYMENT_AMOUNT_ERROR = 6003 // Payment amount error
)

// GetErrorMsg returns error message by error code
func GetErrorMsg(code int) string {
	messages := map[int]string{
		SUCCESS: "success",

		ERROR_SYSTEM:         "System error",
		ERROR_INVALID_PARAMS: "Invalid parameters",
		ERROR_DATABASE:       "Database error",
		ERROR_CACHE:          "Cache error",
		ERROR_RPC:            "RPC call error",
		ERROR_THIRD_PARTY:    "Third-party service error",

		ERROR_USER_NOT_FOUND:      "User not found",
		ERROR_USER_ALREADY_EXISTS: "User already exists",
		ERROR_WRONG_PASSWORD:      "Wrong password",
		ERROR_TOKEN_INVALID:       "Invalid token",
		ERROR_TOKEN_EXPIRED:       "Token expired",
		ERROR_PERMISSION_DENIED:   "Permission denied",

		ERROR_PRODUCT_NOT_FOUND:    "Product not found",
		ERROR_PRODUCT_OUT_OF_STOCK: "Product out of stock",
		ERROR_PRODUCT_INVALID:      "Invalid product",

		ERROR_CART_EMPTY:          "Cart is empty",
		ERROR_CART_ITEM_NOT_FOUND: "Cart item not found",
		ERROR_CART_EXCEED_LIMIT:   "Cart exceeds limit",

		ERROR_ORDER_NOT_FOUND:      "Order not found",
		ERROR_ORDER_STATUS_INVALID: "Invalid order status",
		ERROR_ORDER_CANNOT_CANCEL:  "Cannot cancel order",
		ERROR_ORDER_CREATE_FAILED:  "Order creation failed",

		ERROR_PAYMENT_NOT_FOUND:    "Payment not found",
		ERROR_PAYMENT_FAILED:       "Payment failed",
		ERROR_PAYMENT_TIMEOUT:      "Payment timeout",
		ERROR_PAYMENT_AMOUNT_ERROR: "Payment amount error",
	}

	if msg, ok := messages[code]; ok {
		return msg
	}
	return "Unknown error"
}
