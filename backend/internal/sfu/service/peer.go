package service

import (
	"context"
	"fmt"
	sfu "vidcall/api/proto"
	"vidcall/internal/sfu/domain"
	"vidcall/pkg/logger"

	"github.com/pion/webrtc/v3"
)

type Peer struct {
	*domain.Peer
}

func NewPeer(ctx context.Context, stream sfu.SFU_SignalServer, stuns []string) *Peer {
	// TODO: add Turn server, error handling, and logging

	log := logger.GetLog(ctx).With("layer", "service")

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
				log.Info("no more candidates")
				return
			}

			//  Send ICE candidate
			// TODO: error handling
			cad := c.ToJSON()
			ice := &sfu.IceCandidate{
				Candidate:     cad.Candidate,
				SdpMid:        *cad.SDPMid,
				SdpMlineIndex: uint32(*cad.SDPMLineIndex),
			}

			_ = stream.Send(&sfu.PeerSignal{
				Payload: &sfu.PeerSignal_Ice{
					Ice: ice,
				},
			})

			log.Info("sent ice")
		},
	)

	videoT, _ := pc.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo)
	audioT, _ := pc.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio)

	pc.OnTrack(func(remote *webrtc.TrackRemote, recv *webrtc.RTPReceiver) {
		log.Info(fmt.Sprintf("🔄 got %s – echoing back", remote.Kind()))

		local, _ := webrtc.NewTrackLocalStaticRTP(
			remote.Codec().RTPCodecCapability,
			remote.ID()+"-loop",
			"pion",
		)

		switch remote.Kind() {
		case webrtc.RTPCodecTypeVideo:
			_ = videoT.Sender().ReplaceTrack(local)
		case webrtc.RTPCodecTypeAudio:
			_ = audioT.Sender().ReplaceTrack(local)
		}

		go func() {
			for {
				pkt, _, err := remote.ReadRTP()
				if err != nil {
					log.Error("unable to read RTP")
					return
				}
				_ = local.WriteRTP(pkt)
			}
		}()
	})

	// TODO: renegotiation for subtitles and dubbing

	return &Peer{
		Peer: &domain.Peer{
			Ctx:        ctx,
			PC:         pc,
			IceCanDone: done,
			Stream:     stream,
		},
	}
}

func (p *Peer) Negotiate() {
	log := logger.GetLog(p.Ctx).With("layer", "service")
	ice_buffer := make(chan webrtc.ICECandidateInit, 10)

	for {

		// TODO: handle and log error
		req, err := p.Stream.Recv()
		if err != nil {
			log.Error("something went wrong with stream")
			return
		}

		switch msg := req.GetPayload().(type) {
		// Check on this !
		case *sfu.PeerSignal_Sdp:
			offer := webrtc.SessionDescription{
				Type: webrtc.SDPTypeOffer,
				SDP:  msg.Sdp.Sdp,
			}

			// TODO: error handle
			_ = p.PC.SetRemoteDescription(offer)
			log.Info("set remote description")

			// Flush ice candidate
			go func() {
				for {
					select {
					case c := <-ice_buffer:
						fmt.Println(c)
						p.PC.AddICECandidate(c)
					default:
						return
					}
				}
			}()

			// TODO: error handle and create answer
			answer, _ := p.PC.CreateAnswer(nil)
			p.PC.SetLocalDescription(answer)

			// TODO: error handle and send back SDP
			_ = p.Stream.Send(&sfu.PeerSignal{
				Payload: &sfu.PeerSignal_Sdp{
					Sdp: &sfu.SDP{
						Type: sfu.SdpType_ANSWER,
						Sdp:  p.PC.LocalDescription().SDP,
					},
				},
			})

		case *sfu.PeerSignal_Ice:

			i := uint16(msg.Ice.SdpMlineIndex)
			ice := webrtc.ICECandidateInit{
				Candidate:        msg.Ice.Candidate,
				SDPMid:           &msg.Ice.SdpMid,
				SDPMLineIndex:    &i,
				UsernameFragment: &msg.Ice.UsernameFragment,
			}

			if p.PC.RemoteDescription() != nil {
				if err := p.PC.AddICECandidate(ice); err != nil {
					log.Error(fmt.Sprintf("can't add Ice with error: %v", err))
				}
				log.Info("added ice")
			} else {
				ice_buffer <- ice
			}

		default:

		}
	}
}
