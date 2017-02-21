package events

import (
	"context"

	newrelic "github.com/newrelic/go-agent"
	"github.com/satori/go.uuid"
)

// See https://github.com/golang/lint/pull/245
type contextKey struct {
	name string
}

var NewRelicTransactionCtxKey = &contextKey{"NewRelicTransaction"}
var NewRelicApplicationCtxKey = &contextKey{"NewRelicAplication"}

func RandomID() string {
	return uuid.NewV4().String()
}

// GetNewRelicTransaction get a newrelic transaction from context
func GetNewRelicTransaction(ctx context.Context) newrelic.Transaction {
	if txn, ok := ctx.Value(NewRelicTransactionCtxKey).(newrelic.Transaction); ok {
		return txn
	}

	return nil
}

// ContextWithNewRelicTransaction add  newrelic transaction to the context
func ContextWithNewRelicTransaction(ctx context.Context, txn newrelic.Transaction) context.Context {
	return context.WithValue(ctx, NewRelicTransactionCtxKey, txn)
}

// GetNewRelicApplication get a newrelic application from context
func GetNewRelicApplication(ctx context.Context) newrelic.Application {
	if txn, ok := ctx.Value(NewRelicApplicationCtxKey).(newrelic.Application); ok {
		return txn
	}

	return nil
}
