package domain

import (
	"context"
	"log/slog"
	sfu "vidcall/api/proto"
)

type Peer interface {
	GetMetaData() *PeerMD
	Pub() Publisher
	Sub() Subscriber
	Connect() error
	Disconnect() error
	EnqueueEvent(event *sfu.PeerSignal_Event)
}

type PeerObj struct {
	Metadata   *PeerMD
	Log        *slog.Logger
	Ctx        context.Context
	Cancel     context.CancelFunc
	Stream     sfu.SFU_SignalServer
	Publisher  Publisher
	Subscriber Subscriber

	SendQ  chan *sfu.PeerSignal
	EventQ chan *sfu.PeerSignal_Event
}

type PeerMD struct {
	Name   string
	PeerID string
	RoomID string
	Role   sfu.RoleType
}
