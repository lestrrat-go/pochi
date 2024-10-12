package middleware

import "github.com/lestrrat-go/accesslog"

// AccessLog creates a new `github.com/lestrrat-go/accesslog` middleware.
// This is a simple wrapper around `accesslog.New()`
func AccessLog() *accesslog.Middleware {
	return accesslog.New()
}
