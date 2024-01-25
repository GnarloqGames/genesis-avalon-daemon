package router

import (
	"fmt"
	"log/slog"

	"github.com/GnarloqGames/genesis-avalon-kit/transport"
	"github.com/nats-io/nats.go"
)

type Router struct {
	bus           *transport.Connection
	subscriptions map[string]*nats.Subscription
}

func New(bus *transport.Connection) *Router {
	router := &Router{
		bus:           bus,
		subscriptions: make(map[string]*nats.Subscription),
	}

	cbs := map[string]func(m *nats.Msg){
		"test/simple":  router.HandleSimpleTest,
		"test/request": router.HandleRequestTest,
	}

	for subject, cb := range cbs {
		sub, err := router.bus.Subscribe(subject, cb)

		if err != nil {
			slog.Warn("failed to subscribe", "topic", "test")
			continue
		}

		slog.Info("subscribed to subject", "subject", subject)

		router.subscriptions["test"] = sub
	}

	return router
}

func (r *Router) HandleSimpleTest(m *nats.Msg) {
	fmt.Printf("%+v", m.Data)
}

func (r *Router) HandleRequestTest(m *nats.Msg) {
	fmt.Printf("%+v", m.Data)

	if err := r.bus.Publish(m.Reply, `{"world":"hello"}`); err != nil {
		slog.Warn("Failed to publish response", "error", err.Error(), "subject", m.Reply)
	}
}
