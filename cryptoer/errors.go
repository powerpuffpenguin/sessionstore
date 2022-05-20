package cryptoer

import "errors"

var (
	ErrHashUnavailable        = errors.New(`the requested hash function is unavailable`)
	ErrInvalidToken           = errors.New(`token invalid`)
	ErrNotExistsToken         = errors.New(`token not exists`)
	ErrExpired                = errors.New(`token expired`)
	ErrRefreshTokenNotMatched = errors.New(`refresh token not matched`)
	ErrCannotRefresh          = errors.New(`cannot refresh`)
)
