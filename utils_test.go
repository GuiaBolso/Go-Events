package events

import (
	"context"
	"testing"

	newrelic "github.com/newrelic/go-agent"
)

func Test_GetNewRelicTransaction_Success(t *testing.T) {
	newrelicConfig := newrelic.NewConfig("AppName", "")
	newrelicConfig.Enabled = false

	app, _ := newrelic.NewApplication(newrelicConfig)
	txn := app.StartTransaction("transaction_name", nil, nil)

	ctx := ContextWithNewRelicTransaction(context.Background(), txn)

	if GetNewRelicTransaction(ctx) == nil {
		t.Error("Must return a newrelic transaction")
	}
}

func Test_GetNewRelicTransaction_Failure(t *testing.T) {
	ctx := ContextWithNewRelicTransaction(context.Background(), nil)

	if GetNewRelicTransaction(ctx) != nil {
		t.Error("Must return a nil transaction")
	}
}
