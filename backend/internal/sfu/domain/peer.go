package domain

import (
	"log/slog"
	"sync"
	sfu "vidcall/api/proto"

	"github.com/pion/webrtc/v3"
)

type Peer struct {
	ID         string
	Log        *slog.Logger
	Stream     sfu.SFU_SignalServer
	Publisher  *Publisher
	Subscriber *Subscriber

	SendQ chan *sfu.PeerSignal
}

type Publisher struct {
	PC         *webrtc.PeerConnection
	LocalAudio *webrtc.TrackLocalStaticRTP
	LocalVideo *webrtc.TrackLocalStaticRTP
}

type Subscriber struct {
	Mu           sync.RWMutex
	PC           *webrtc.PeerConnection
	IdleAudios   []*webrtc.RTPTransceiver
	IdleVideos   []*webrtc.RTPTransceiver
	Dub          *webrtc.RTPTransceiver
	Sub          *webrtc.DataChannel
	ActiveTracks map[*webrtc.TrackLocalStaticRTP]*webrtc.RTPTransceiver
}
