package domain

import (
	"log/slog"
	"sync"
	"time"
	sfu "vidcall/api/proto"

	"github.com/pion/webrtc/v3"
)

type Connection interface {
	GetPC() *webrtc.PeerConnection
	GetAudioURI() string
	HandleLocalIce(candidate *webrtc.ICECandidate, pc sfu.PcType)
	HandleRemoteIce(candidate *sfu.PeerSignal_Ice) error
	HandleOffer(sdp *sfu.PeerSignal_Sdp) error
	HandleAnswer(sdp *sfu.PeerSignal_Sdp) error
	SendOffer(pc sfu.PcType) error
	Close() error
}

type PConn struct {
	PC            *webrtc.PeerConnection
	AudioLevelURI string
	Log           *slog.Logger
	IceBuffers    chan webrtc.ICECandidateInit
	RecvQ         chan *sfu.PeerSignal
	SendQ         chan *sfu.PeerSignal
	DebounceTimer *DebounceTimer
}

type DebounceTimer struct {
	Mu       sync.Mutex
	Timer    *time.Timer
	Interval time.Duration
}
