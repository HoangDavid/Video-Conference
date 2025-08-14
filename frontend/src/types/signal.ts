
export type Signal =
    | {type: "sdp", payload: Sdp}
    | {type: "ice", payload: Ice}
    | {type: "action", payload: PeerAction}
    | {type: "event", payload: PeerEvent}

export type SdpType = "offer" | "answer"
export type PcType = "pub" | "sub" | "pc_unspecified"
export type RoleType = "host" | "guest" | "bot" | "role_unspecified"
export type ActionType = "start_room" | "end_room" | "join" | "leave" | "audio_on" | "audio_off" | 
        "video_on" | "video_off" | "dubbing_on" | "dubbing_off"
export type EventType = "room_active" | "room_inactive" | "room_ended" | "join_event" | "leave_event" |
        "audio_enabled" | "audio_disabled" | "video_enabled" | "video_disabled"


export interface Sdp{
    pc: PcType
    type: SdpType
    sdp: string
}

export interface Ice {
    pc: PcType
    candidate: string
    sdpMid: string 
	sdpMLineIndex: number
	usernameFragment: string
}

export interface PeerAction {
    peerID: string
    roomID: string
    type: ActionType
    role: RoleType
}

export interface PeerEvent {
    peerID: string
    roomID: string
    type: EventType
}