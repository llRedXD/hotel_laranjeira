package auth

import (
	"context"
	"time"
)

// Repository accepts only password/token hashes. Multi-write operations must
// commit atomically; failed rotation must leave the previous token usable.
type Repository interface {
	InsertUser(context.Context, string, string, string, string) error
	UserByEmail(context.Context, string) (User, string, error)
	ActiveUser(context.Context, string, int64) (User, error)
	InsertRefresh(context.Context, string, string, time.Time) error
	RotateRefresh(context.Context, string, string, time.Time) (User, error)
	RevokeRefresh(context.Context, string) error
	ReplaceReset(context.Context, string, string, time.Time) error
	ConsumeReset(context.Context, string, string) error
}
