package wsx

import (
	"encoding/json"
	sfu "vidcall/api/proto"
)

func handleClientSDP(payload json.RawMessage) (*sfu.PeerSignal, error) {
	var (
		sdp     sdp
		pcType  sfu.PcType
		sdpType sfu.SdpType
		signal  *sfu.PeerSignal
	)

	if err := json.Unmarshal(payload, &sdp); err != nil {
		return nil, err
	}

	switch sdp.Pc {
	case "pub":
		pcType = sfu.PcType_PUB
	case "sub":
		pcType = sfu.PcType_SUB
	default:
		pcType = sfu.PcType_PC_UNSPECIFIED
	}

	switch sdp.Type {
	case "offer":
		sdpType = sfu.SdpType_OFFER

	case "answer":
		sdpType = sfu.SdpType_ANSWER

	}

	signal = &sfu.PeerSignal{
		Payload: &sfu.PeerSignal_Sdp{
			Sdp: &sfu.Sdp{
				Pc:   pcType,
				Type: sdpType,
				Sdp:  sdp.SDP,
			},
		},
	}

	return signal, nil
}

func handleClientIce(payload json.RawMessage) (*sfu.PeerSignal, error) {
	var (
		ice    ice
		pcType sfu.PcType
		signal *sfu.PeerSignal
	)

	if err := json.Unmarshal(payload, &ice); err != nil {
		return nil, err
	}

	switch ice.Pc {
	case "pub":
		pcType = sfu.PcType_PUB
	case "sub":
		pcType = sfu.PcType_SUB
	default:
		pcType = sfu.PcType_PC_UNSPECIFIED
	}

	signal = &sfu.PeerSignal{
		Payload: &sfu.PeerSignal_Ice{
			Ice: &sfu.IceCandidate{
				Pc:               pcType,
				Candidate:        ice.Candidate,
				SdpMid:           ice.SdpMid,
				SdpMlineIndex:    ice.SdpMLineIndex,
				UsernameFragment: ice.UsernameFragmment,
			},
		},
	}

	return signal, nil
}

func handleClientAction(payload json.RawMessage) (*sfu.PeerSignal, error) {
	var (
		action  action
		actType sfu.ActionType
		signal  *sfu.PeerSignal
	)

	if err := json.Unmarshal(payload, &action); err != nil {
		return nil, err
	}

	switch action.Type {
	case "start_room":
		actType = sfu.ActionType_START_ROOM
	case "join":
		actType = sfu.ActionType_JOIN
	case "leave":
		actType = sfu.ActionType_LEAVE
	case "end_room":
		actType = sfu.ActionType_END_ROOM
	}

	signal = &sfu.PeerSignal{
		Payload: &sfu.PeerSignal_Action{
			Action: &sfu.Action{
				Peerid: action.PeerID,
				Roomid: action.RoomID,
				Type:   actType,
			},
		},
	}

	return signal, nil
}
