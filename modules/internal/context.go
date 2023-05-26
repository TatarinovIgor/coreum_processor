package internal

import (
	"context"
	"coreum_processor/modules/storage"
)

type key int

const (
	keyExternalID key = iota
	keyMerchantID
	keyUserStore
)

func WithExternalID(ctx context.Context, externalID string) context.Context {
	return context.WithValue(ctx, keyExternalID, externalID)
}

func WithMerchantID(ctx context.Context, merchantID string) context.Context {
	return context.WithValue(ctx, keyMerchantID, merchantID)
}

func WithUserStore(ctx context.Context, userStore *storage.UserStore) context.Context {
	return context.WithValue(ctx, keyUserStore, userStore)
}

func GetExternalID(ctx context.Context) (string, error) {
	return getStringValue(ctx, keyExternalID)
}

func GetMerchantID(ctx context.Context) (string, error) {
	return getStringValue(ctx, keyMerchantID)
}

func GetUserStore(ctx context.Context) (*storage.UserStore, error) {
	valueRaw := ctx.Value(keyUserStore)
	if valueRaw == nil {
		return nil, ErrNotFound
	}
	value, ok := valueRaw.(*storage.UserStore)
	if !ok {
		return nil, ErrTypeMismatch
	}
	return value, nil
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

func getBoolValue(ctx context.Context, k key) (bool, error) {
	valueRaw := ctx.Value(k)
	if valueRaw == nil {
		return false, ErrNotFound
	}
	value, ok := valueRaw.(bool)
	if !ok {
		return false, ErrTypeMismatch
	}
	return value, nil
}
