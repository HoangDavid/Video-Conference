package service

import (
	"fmt"
	sfu "vidcall/api/proto"
	"vidcall/internal/sfu/service/hub"
	"vidcall/internal/sfu/service/room"
)

func (p *PeerObj) handleActions(act *sfu.PeerSignal_Action) error {
	md := p.Metadata
	r := hub.Hub().GetRoom(md.RoomID)

	if r == nil {
		r = room.NewRoom(md.RoomID)
	}

	log := p.Log.With("handlers", "action", "peer ID", md.PeerID)

	switch act.Action.Type {
	case sfu.ActionType_START_ROOM:
		if r.GetPeer(md.PeerID) == nil {
			r.AddPeer(md.PeerID, p)
		}

		if !r.IsLive() && md.Role == sfu.RoleType_ROLE_HOST {
			r.MakeLive()

			roomActiveE := p.createEvent(md.RoomID, sfu.EventType_ROOM_ACTIVE)
			r.BroadCast(md.PeerID, roomActiveE)

			log.Info("host start room")
		}

		return nil

	case sfu.ActionType_JOIN:
		fmt.Println(md.PeerID, "joining")
		if r.GetPeer(md.PeerID) == nil {
			r.AddPeer(md.PeerID, p)
		}

		// room is not live
		if !r.IsLive() {
			roomInactiveE := p.createEvent(md.RoomID, sfu.EventType_ROOM_INACTIVE)
			p.EnqueueSend(&sfu.PeerSignal{Payload: roomInactiveE})
			return nil
		} else {
			roomActiveE := p.createEvent(md.RoomID, sfu.EventType_ROOM_ACTIVE)
			p.EnqueueSend(&sfu.PeerSignal{Payload: roomActiveE})
		}
		if err := p.Subscriber.SubscribeRoom(md.PeerID, r); err != nil {
			return err
		}
		fmt.Println(md.PeerID, "subcribed to room")
		// create event and broadcast
		joinE := p.createEvent(md.RoomID, sfu.EventType_JOIN_EVENT)
		r.BroadCast(md.PeerID, joinE)
		log.Info("guest join room")

		return nil

	case sfu.ActionType_LEAVE:
		if r.GetPeer(md.PeerID) != nil {
			r.RemovePeer(md.PeerID)
		}

		if r.IsLive() {
			// create event and broadcast
			leaveE := p.createEvent(md.RoomID, sfu.EventType_LEAVE_EVENT)
			r.BroadCast(md.PeerID, leaveE)

			// trigger context to disconnect pc
			p.Cancel()
			log.Info("guest leave room")
		}

		return nil

	case sfu.ActionType_END_ROOM:
		if r.IsLive() && md.Role == sfu.RoleType_ROLE_HOST {
			r.Close()
			// create end room event
			endRoomE := p.createEvent(md.RoomID, sfu.EventType_ROOM_ENDED)
			r.BroadCast(md.PeerID, endRoomE)

			// trigger to disconnect pc
			p.Cancel()
			log.Info("Action: end room")

			return nil
		}

	case sfu.ActionType_AUDIO_ON:
		if r.IsLive() {
			audioOnE := p.createEvent(md.RoomID, sfu.EventType_AUDIO_ENABLED)
			r.BroadCast(md.PeerID, audioOnE)

			log.Info("Action: audio enabled")
		}
	case sfu.ActionType_AUDIO_OFF:
		if r.IsLive() {
			audioOffE := p.createEvent(md.RoomID, sfu.EventType_AUDIO_DISABLED)
			r.BroadCast(md.PeerID, audioOffE)

			log.Info("Action: audio disabled")
		}
	case sfu.ActionType_VIDEO_ON:
		if r.IsLive() {
			videoOnE := p.createEvent(md.RoomID, sfu.EventType_VIDEO_ENABLED)
			r.BroadCast(md.PeerID, videoOnE)

			log.Info("Action: video enabled")
		}

	case sfu.ActionType_VIDEO_OFF:
		if r.IsLive() {
			videoOffE := p.createEvent(md.RoomID, sfu.EventType_VIDEO_DISABLED)
			r.BroadCast(md.PeerID, videoOffE)

			log.Info("Action: video disabled")
		}
	case sfu.ActionType_DUBBING_ON:
	case sfu.ActionType_DUBBING_OFF:
	default:
	}

	return nil
}

func (p *PeerObj) handleEvents(evt *sfu.PeerSignal_Event) error {

	md := p.Metadata

	log := p.Log.With("handlers", "event", "peer ID", md.PeerID)

	switch evt.Event.Type {
	case sfu.EventType_ROOM_ACTIVE:
		p.EnqueueSend(&sfu.PeerSignal{Payload: evt})
		log.Info("room active event")

	case sfu.EventType_ROOM_INACTIVE:
		p.EnqueueSend(&sfu.PeerSignal{Payload: evt})
		log.Info("room inactive event")

	case sfu.EventType_JOIN_EVENT:
		r := hub.Hub().GetRoom(md.RoomID)
		_ = r.GetPeer(md.PeerID)

		if evt.Event.PeerID != md.PeerID {
			peer := r.GetPeer(evt.Event.PeerID)
			if peer == nil {
				p.Log.Error("peer does not exist")
				return nil
			}

			if err := p.Subscriber.Subscribe(peer); err != nil {
				return err
			}

			fmt.Println(md.PeerID, "subcribed to ", peer.GetMetaData().PeerID)
		}

		p.EnqueueSend(&sfu.PeerSignal{Payload: evt})
		log.Info("peer join event")

	case sfu.EventType_LEAVE_EVENT:

		if evt.Event.PeerID != md.PeerID {
			if err := p.Subscriber.Unsubscribe(evt.Event.PeerID); err != nil {
				return err
			}
		}

		p.EnqueueSend(&sfu.PeerSignal{Payload: evt})
		log.Info("peer leave event")

	case sfu.EventType_ROOM_ENDED:
		p.Cancel()
		log.Info("host end room event")

	case sfu.EventType_AUDIO_ENABLED:
		p.EnqueueSend(&sfu.PeerSignal{Payload: evt})
		log.Info("audio enabled event")
	case sfu.EventType_AUDIO_DISABLED:
		p.EnqueueSend(&sfu.PeerSignal{Payload: evt})
		log.Info("audio disabled event")
	case sfu.EventType_VIDEO_ENABLED:
		p.EnqueueSend(&sfu.PeerSignal{Payload: evt})
		log.Info("video enabled event")
	case sfu.EventType_VIDEO_DISABLED:
		p.EnqueueSend(&sfu.PeerSignal{Payload: evt})
		log.Info("video disabled event")
	default:
	}
	return nil
}

// helper funciton to create event
func (p *PeerObj) createEvent(_ string, e sfu.EventType) *sfu.PeerSignal_Event {
	return &sfu.PeerSignal_Event{
		Event: &sfu.Event{
			Name:   p.Metadata.Name,
			PeerID: p.Metadata.PeerID,
			Type:   e,
		},
	}
}
