package internal

import "context"

type key int

const (
	keyExternalID key = iota
	keyMerchantID
)

func WithExternalID(ctx context.Context, externalID string) context.Context {
	return context.WithValue(ctx, keyExternalID, externalID)
}

func WithMerchantID(ctx context.Context, merchantID string) context.Context {
	return context.WithValue(ctx, keyMerchantID, merchantID)
}

func GetExternalID(ctx context.Context) (string, error) {
	return getStringValue(ctx, keyExternalID)
}

func GetMerchantID(ctx context.Context) (string, error) {
	return getStringValue(ctx, keyMerchantID)
}

func getStringValue(ctx context.Context, k key) (string, error) {
	valueRaw := ctx.Value(k)
	if valueRaw == nil {
		return "", ErrNotFound
	}
	value, ok := valueRaw.(string)
	if !ok {
		return "", ErrTypeMismatch
	}
	return value, nil
}
