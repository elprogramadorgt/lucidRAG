// Package common provides shared domain types used across multiple domains.
package common

// UserContext contains user identity and authorization information
// for request-scoped operations.
type UserContext struct {
	UserID  string
	IsAdmin bool
}
