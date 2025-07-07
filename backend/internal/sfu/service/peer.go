package service

import (
	"context"
	sfu "vidcall/api/proto"
	"vidcall/internal/sfu/domain"

	"github.com/pion/webrtc/v3"
)

type Peer struct {
	*domain.Peer
}

func NewPeer(ctx context.Context, stream sfu.SFU_SignalServer, stuns []string) *Peer {
	// TODO: add Turn server, error handling, and logging
	pc, _ := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	})

	done := make(chan struct{})
	// TODO: on network failure
	pc.OnICECandidate(
		func(c *webrtc.ICECandidate) {
			if c == nil {
				close(done)
				return
			}

			//  Send ICE candidate
			// TODO: error handling
			_ = stream.Send(&sfu.PeerResponse{
				Payload: &sfu.PeerResponse_Ice{
					Ice: &sfu.IceCandidate{
						Candidate: c.ToJSON().Candidate,
					},
				},
			})
		},
	)

	return &Peer{
		Peer: &domain.Peer{
			PC:         pc,
			IceCanDone: done,
			Stream:     stream,
		},
	}
}

func (p *Peer) Negotiate() {
	for {
		req, _ := p.Stream.Recv()

		switch msg := req.GetPayload().(type) {
		// Check on this !
		case *sfu.PeerRequest_Offer:
			offer := webrtc.SessionDescription{
				Type: webrtc.SDPTypeOffer,
				SDP:  msg.Offer.Sdp,
			}

			// TODO: error handle
			_ = p.PC.SetRemoteDescription(offer)

			// TODO: error handle and create answer
			answer, _ := p.PC.CreateAnswer(nil)
			p.PC.SetLocalDescription(answer)

			// TODO: error handle and send back SDP
			<-webrtc.GatheringCompletePromise(p.PC)
			_ = p.Stream.Send(&sfu.PeerResponse{
				Payload: &sfu.PeerResponse_Answer{
					Answer: &sfu.SDP{
						Type: sfu.SdpType_ANSWER,
						Sdp:  p.PC.LocalDescription().SDP,
					},
				},
			})

		case *sfu.PeerRequest_Ice:
			_ = p.PC.AddICECandidate(webrtc.ICECandidateInit{
				Candidate: msg.Ice.Candidate,
			})

		default:

		}
	}
}
