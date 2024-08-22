package router

import (
	"log/slog"
	"time"

	"github.com/GnarloqGames/genesis-avalon-daemon/worker"
	"github.com/GnarloqGames/genesis-avalon-kit/proto"
	"github.com/GnarloqGames/genesis-avalon-kit/transport"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	goproto "google.golang.org/protobuf/proto"
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

func (r *Router) HandleBuild(b *nats.Msg) {
	var data proto.BuildRequest

	if err := goproto.Unmarshal(b.Data, &data); err != nil {
		slog.Error("failed to parse task", "error", err.Error())

		res := &proto.BuildResponse{
			Header: &proto.ResponseHeader{
				Timestamp: timestamppb.Now(),
				Status:    proto.Status_ERROR,
			},
			Response: "failed to parse task",
		}

		if err := transport.Respond(r.bus, b.Reply, res); err != nil {
			slog.Warn("Failed to publish response", "error", err.Error(), "subject", b.Reply)
		}

		return
	}

	dur, err := time.ParseDuration(data.Duration)
	if err != nil {
		slog.Error("failed to parse task duration", "duration", data.Duration)

		res := &proto.BuildResponse{
			Header: &proto.ResponseHeader{
				Timestamp: timestamppb.Now(),
				Status:    proto.Status_ERROR,
			},
			Response: "failed to parse task duration",
		}

		if err := transport.Respond(r.bus, b.Reply, res); err != nil {
			slog.Warn("Failed to publish response", "error", err.Error(), "subject", b.Reply)
		}

		return
	}

	ownerField, ok := data.Context.Fields["owner"]
	if !ok {
		slog.Error("owner value missing")

		res := &proto.BuildResponse{
			Header: &proto.ResponseHeader{
				Timestamp: timestamppb.Now(),
				Status:    proto.Status_ERROR,
			},
			Response: "owner value missing",
		}

		if err := transport.Respond(r.bus, b.Reply, res); err != nil {
			slog.Warn("Failed to publish response", "error", err.Error(), "subject", b.Reply)
		}
	}

	r.pool.Inbox() <- &worker.BuildTask{
		ID:       uuid.New(),
		Name:     data.Name,
		Duration: dur,
		Owner:    ownerField.GetStringValue(),
	}

	res := &proto.BuildResponse{
		Header: &proto.ResponseHeader{
			Timestamp: timestamppb.Now(),
			Status:    proto.Status_OK,
		},
		Response: "building queued",
	}

	if err := transport.Respond(r.bus, b.Reply, res); err != nil {
		slog.Warn("Failed to publish response", "error", err.Error(), "subject", b.Reply)
	}
}
