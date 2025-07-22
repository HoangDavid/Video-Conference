package service

import (
	sfu "vidcall/api/proto"
	"vidcall/internal/sfu/service/hub"
	"vidcall/internal/sfu/service/room"
)

func (p *PeerObj) handleStartRoom(roomID string) {
	// create room instance
	r := room.NewRoom(roomID)
	r.AddPeer(p.ID, p)

	// create room active event
	roomActiveE := &sfu.PeerSignal_Event{
		Event: &sfu.Event{
			Roomid: roomID,
			Peerid: p.ID,
			Type:   sfu.EventType_ROOM_ACTIVE,
		},
	}

	// broadcast to peers in lobby
	r.BroadCast(p.ID, roomActiveE)

}

func (p *PeerObj) handleJoinRoom(roomID string) {

	r := hub.Hub().GetRoom(roomID)

	// room is not live
	if r == nil {
		roomInactiveE := &sfu.PeerSignal_Event{
			Event: &sfu.Event{
				Roomid: roomID,
				Peerid: p.ID,
				Type:   sfu.EventType_ROOM_INACTIVE,
			},
		}

		p.EnqueueEvent(roomInactiveE)
	}

	// add peer to live room
	r.AddPeer(p.ID, p)

	// create join event
	joinE := &sfu.PeerSignal_Event{
		Event: &sfu.Event{
			Roomid: roomID,
			Peerid: p.ID,
			Type:   sfu.EventType_JOIN_EVENT,
		},
	}

	r.BroadCast(p.ID, joinE)

}

func (p *PeerObj) handleLeaveRoom(roomID string) {
	r := hub.Hub().GetRoom(roomID)
	r.RemovePeer(p.ID)

	// stop receiving a/v from other users
	p.Subscriber.UnsubscribeRoom(p.ID, r)

	// create leave event
	leaveE := &sfu.PeerSignal_Event{
		Event: &sfu.Event{
			Roomid: roomID,
			Peerid: p.ID,
			Type:   sfu.EventType_LEAVE_EVENT,
		},
	}

	r.BroadCast(p.ID, leaveE)

	// stop recieving requests from signaling
	p.Cancel()
}

func (p *PeerObj) handleEndRoom(roomID string) {
	r := hub.Hub().RemoveRoom(roomID)

	// create end room event
	endRoomE := &sfu.PeerSignal_Event{
		Event: &sfu.Event{
			Roomid: roomID,
			Peerid: p.ID,
			Type:   sfu.EventType_ROOM_ENEDED,
		},
	}

	r.BroadCast(p.ID, endRoomE)

	p.Cancel()
}

func (p *PeerObj) handleRoomActiveEvent(event *sfu.PeerSignal_Event) {
	roomActiveE := &sfu.PeerSignal{
		Payload: event,
	}

	p.SendQ <- roomActiveE
}

func (p *PeerObj) handleRoomInactiveEvent(event *sfu.PeerSignal_Event) {
	roomInactiveE := &sfu.PeerSignal{
		Payload: event,
	}

	p.SendQ <- roomInactiveE
}

func (p *PeerObj) handleJoinEvent(event *sfu.PeerSignal_Event) error {
	peerID := event.Event.Peerid
	roomID := event.Event.Roomid

	r := hub.Hub().GetRoom(roomID)
	peer := r.GetPeer(peerID)

	if err := p.Subscriber.Subscribe(peer); err != nil {
		return err
	}

	return nil
}

func (p *PeerObj) handleLeaveEvent(event *sfu.PeerSignal_Event) error {
	peerID := event.Event.Peerid
	roomID := event.Event.Roomid

	r := hub.Hub().GetRoom(roomID)
	r.RemovePeer(peerID)

	if err := p.Subscriber.Unsubscribe(peerID); err != nil {
		return err
	}

	return nil
}

func (p *PeerObj) handleRoomEndedEvent() {
	p.Cancel()
}
