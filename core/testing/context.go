package coretesting

import (
	"context"

	"cosmossdk.io/core/event"
	"cosmossdk.io/core/store"
	"google.golang.org/protobuf/runtime/protoiface"
)

type dummyKey struct{}

func Context() context.Context {
	dummy := &dummyCtx{
		stores:      map[string]store.KVStore{},
		events:      map[string][]event.Event{},
		protoEvents: map[string][]protoiface.MessageV1{},
	}

	ctx := context.WithValue(context.Background(), dummyKey{}, dummy)
	return ctx
}

type dummyCtx struct {
	// maps store by the actor.
	stores map[string]store.KVStore
	// maps event emitted by the actor.
	events map[string][]event.Event
	// maps proto events emitted by the actor.
	protoEvents map[string][]protoiface.MessageV1
}

func unwrap(ctx context.Context) *dummyCtx {
	dummy := ctx.Value(dummyKey{})
	if dummy == nil {
		panic("invalid ctx without dummy")
	}

	return dummy.(*dummyCtx)
}
