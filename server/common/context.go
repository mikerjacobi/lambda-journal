package common

import "context"

const (
	ctxContentType = iota
)

func WithContentType(ctx context.Context, contentType string) context.Context {
	return context.WithValue(ctx, ctxContentType, contentType)
}
