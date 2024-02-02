package router

import (
	"log/slog"

	"github.com/GnarloqGames/genesis-avalon-daemon/worker"
	"github.com/GnarloqGames/genesis-avalon-kit/proto"
	"github.com/GnarloqGames/genesis-avalon-kit/transport"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Router struct {
	bus           *transport.Connection
	pool          *worker.System
	subscriptions map[string]*nats.Subscription
}

func New(bus *transport.Connection, pool *worker.System) *Router {
	router := &Router{
		bus:           bus,
		pool:          pool,
		subscriptions: make(map[string]*nats.Subscription),
	}

	sub, err := router.bus.Subscribe("build", router.HandleBuild)
	if err != nil {
		slog.Warn("failed to subscribe", "topic", "test")
	}
	router.subscriptions["test"] = sub

	return router
}

func (r *Router) HandleBuild(subj, reply string, b *proto.BuildRequest) {
	res := &proto.BuildResponse{
		Header: &proto.ResponseHeader{
			Timestamp: timestamppb.Now(),
			Status:    proto.Status_OK,
		},
		Response: "hi",
	}

	if err := r.bus.Publish(reply, res); err != nil {
		slog.Warn("Failed to publish response", "error", err.Error(), "subject", reply)
	}
}
