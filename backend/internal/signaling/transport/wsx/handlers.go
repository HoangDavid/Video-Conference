package wsx

import (
	"encoding/json"
	"log/slog"
	sfu "vidcall/api/proto"

	"github.com/gorilla/websocket"
)

func handleClientSDP(payload json.RawMessage, log *slog.Logger) (*sfu.PeerSignal, error) {
	var (
		sdp     sdp
		pcType  sfu.PcType
		sdpType sfu.SdpType
	)

	if err := json.Unmarshal(payload, &sdp); err != nil {
		log.Error("unable to unmarshal the sdp payload")
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

	signal := &sfu.PeerSignal{
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

func handleClientIce(payload json.RawMessage, log *slog.Logger) (*sfu.PeerSignal, error) {
	var (
		ice    ice
		pcType sfu.PcType
	)

	if err := json.Unmarshal(payload, &ice); err != nil {
		log.Error("unable to unmarshal the ice payload")
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

	signal := &sfu.PeerSignal{
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

func handleClientAction(payload json.RawMessage, log *slog.Logger) (*sfu.PeerSignal, error) {
	var (
		action  action
		actType sfu.ActionType
	)

	if err := json.Unmarshal(payload, &action); err != nil {
		log.Error("unable to unmarshal action payload")
		return nil, err
	}

	switch action.Type {
	case "start_room":
		actType = sfu.ActionType_START_ROOM
		log.Info("start room")
	case "join":
		actType = sfu.ActionType_JOIN
		log.Info("join room")
	case "leave":
		actType = sfu.ActionType_LEAVE
		log.Info("leave room")
	case "end_room":
		actType = sfu.ActionType_END_ROOM
		log.Info("end room")
	case "audio_on":
		actType = sfu.ActionType_AUDIO_ON
		log.Info("audio on")
	case "audio_off":
		actType = sfu.ActionType_AUDIO_OFF
		log.Info("audio off")
	case "video_on":
		actType = sfu.ActionType_VIDEO_ON
		log.Info("video on")
	case "video_off":
		actType = sfu.ActionType_VIDEO_OFF
		log.Info("video off")
	case "dubbing_on":
		actType = sfu.ActionType_DUBBING_ON
		log.Info("dubbing on")
	case "dubbing_off":
		actType = sfu.ActionType_DUBBING_OFF
		log.Info("dubbing off")
	}

	signal := &sfu.PeerSignal{
		Payload: &sfu.PeerSignal_Action{
			Action: &sfu.Action{
				Type: actType,
			},
		},
	}

	return signal, nil
}

func handleSfuSDP(msg *sfu.PeerSignal_Sdp, log *slog.Logger) (*signal, error) {
	var (
		pcType  string
		sdpType string
	)

	switch msg.Sdp.Pc {
	case sfu.PcType_PUB:
		pcType = "pub"
	case sfu.PcType_SUB:
		pcType = "sub"
	case sfu.PcType_PC_UNSPECIFIED:
		pcType = "unspecified"
	}

	switch msg.Sdp.Type {
	case sfu.SdpType_OFFER:
		sdpType = "offer"
	case sfu.SdpType_ANSWER:
		sdpType = "answer"
	}

	sdp := sdp{
		Pc:   pcType,
		Type: sdpType,
		SDP:  msg.Sdp.Sdp,
	}

	raw, err := json.Marshal(sdp)
	if err != nil {
		log.Error("unable to marshal sdp payload")
		return nil, err
	}

	s := &signal{
		Type:    "sdp",
		Payload: raw,
	}

	return s, nil

}

func handleSfuIce(msg *sfu.PeerSignal_Ice, log *slog.Logger) (*signal, error) {
	var pcType string

	switch msg.Ice.Pc {
	case sfu.PcType_PUB:
		pcType = "pub"
	case sfu.PcType_SUB:
		pcType = "sub"
	case sfu.PcType_PC_UNSPECIFIED:
		pcType = "unspecified"
	}

	ice := ice{
		Pc:                pcType,
		Candidate:         msg.Ice.Candidate,
		SdpMid:            msg.Ice.SdpMid,
		SdpMLineIndex:     msg.Ice.SdpMlineIndex,
		UsernameFragmment: msg.Ice.UsernameFragment,
	}

	raw, err := json.Marshal(ice)
	if err != nil {
		log.Error("unable to marshal ice payload")
		return nil, err
	}

	s := &signal{
		Type:    "ice",
		Payload: raw,
	}

	return s, nil

}

func handleSfuEvent(msg *sfu.PeerSignal_Event, log *slog.Logger) (*signal, error) {
	var eventType string

	switch msg.Event.Type {
	case sfu.EventType_ROOM_ACTIVE:
		eventType = "room_active"
		log.Info("room active")

	case sfu.EventType_ROOM_INACTIVE:
		eventType = "room_inactive"
		log.Info("room inactive")

	case sfu.EventType_ROOM_ENDED:
		eventType = "room_ended"
		log.Info("room ended")

	case sfu.EventType_JOIN_EVENT:
		eventType = "join_event"
		log.Info("join event")

	case sfu.EventType_LEAVE_EVENT:
		eventType = "leave_event"
		log.Info("leave event")

	case sfu.EventType_AUDIO_ENABLED:
		eventType = "audio_enabled"
		log.Info("audio enabled")

	case sfu.EventType_AUDIO_DISABLED:
		eventType = "audio_disabled"
		log.Info("audio disabled")

	case sfu.EventType_VIDEO_ENABLED:
		eventType = "video_enabled"
		log.Info("video enabled")

	case sfu.EventType_VIDEO_DISABLED:
		eventType = "video_disabled"
		log.Info("video disabled")
	}

	event := event{
		Name:   msg.Event.Name,
		PeerID: msg.Event.PeerID,
		Type:   eventType,
	}

	raw, err := json.Marshal(event)
	if err != nil {
		log.Error("unable to marshal event payload")
		return nil, err
	}

	s := &signal{
		Type:    "event",
		Payload: raw,
	}

	return s, nil
}

type Intent int

const (
	IntentUnknown Intent = iota
	IntentJoin
	IntentExit
)

func handleFirstMsg(conn *websocket.Conn, log *slog.Logger) (Intent, *sfu.PeerSignal, error) {
	var msg signal
	for {
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Error("unable to read msg")
			return IntentUnknown, nil, err
		}

		switch msg.Type {
		case "action":
			pl := msg.Payload
			var act action
			if err := json.Unmarshal(pl, &act); err != nil {
				log.Error("unable to unmarshal")
				return IntentUnknown, nil, err
			}

			switch act.Type {
			case "start_room", "join":
				first, err := handleClientAction(msg.Payload, log)
				if err != nil {
					return IntentUnknown, nil, err
				}

				return IntentJoin, first, nil
			case "leave", "end_room":
				return IntentExit, nil, nil
			default:
			}

		default:
		}
	}
}
