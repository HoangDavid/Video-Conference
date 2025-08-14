import type {Claims}  from "../types/claims";
import type {Sdp, Ice, PcType, PeerEvent} from "../types/signal";
import { RtcClient } from "./lib/rtc";
import { initSignal, getSignal, initRtcPub, initRtcSub } from "./lib/instance";
import { router } from "../main";

export default async function join_meeting(): Promise<boolean>{
    const res = await fetch("/api/me",  {
            method: "GET",
            credentials: "include"
        });

    if (!res.ok) {
        const msg = await res.text().catch(() => "");
        console.log(`HTTP ${res.status} ${msg}`)
        return false
    }

    const claims = await res.json() as Claims

    //  Establish websocket connection
    const signal_conn = initSignal("/ws");

    // Establish peer connection with 
    const pub_conn = initRtcPub();
    const sub_conn = initRtcSub();

    //  Attach callbacks to ws conn
    signal_conn.onSdp((sdp: Sdp) => {
            HandleRemoteSdp(sdp, pub_conn, sdp.pc);
    })

    signal_conn.onIce((ice: Ice) => {
        if (ice.pc == "pub") {
            HandleRemoteIce(ice, pub_conn);
        }else if (ice.pc == "sub") {
            HandleRemoteIce(ice, sub_conn);
        }
    })

    // del: remove callback
    const del = signal_conn.onEvent((e: PeerEvent) => {
        if (e.type == "room_inactive"){
            router.navigate(`/rooms/${e.roomID}/lobby`)
            del();
        } else if (e.type == "room_active") {
            router.navigate(`/rooms/${e.roomID}/meeting`)
            del();
        }
    })

    //  Attach callbacks to peer conn
    pub_conn.onIce((ice: RTCIceCandidate) => {
        signal_conn.sendIce("pub", ice);
    })

    sub_conn.onIce((ice: RTCIceCandidate) => {
        signal_conn.sendIce("sub", ice);
    })

    // Create offer for publisher pc
    const offer = await pub_conn.createOfferAndSetLocal();
    signal_conn.sendSdp("pub","offer", offer);

    // Open websocket
    signal_conn.connect();

    if (claims.role == 'host') {
        //  TODO: send a start meeting request
        signal_conn.sendAction(claims, "start_room");
    }else if (claims.role == 'guest') {
        //  TODO: send a join meeting request: if room not active -> lobby else go straight to room
        signal_conn.sendAction(claims, "join");
    }


    return true
}

// Handle sdp from remote peer
async function HandleRemoteSdp(sdp: Sdp, p_conn: RtcClient, pcType: PcType) {
    if (sdp.type == "answer") {
        await p_conn.setRemoteAnswer(sdp.sdp);
    }else if (sdp.type == "offer") {
        const answer = await p_conn.answerRemoteOffer(sdp.sdp);
        if (!answer) {return}
        const signal_conn = getSignal()
        signal_conn?.sendSdp(pcType, "answer", answer)
    }
}


// Handle Ice from remote peer
function HandleRemoteIce(ice: Ice, p_conn: RtcClient) {
    const ice_can: RTCIceCandidateInit = {
        candidate: ice.candidate,
        sdpMid: ice.sdpMid,
        sdpMLineIndex: ice.sdpMLineIndex,
    };

    p_conn.setRemoteIce(ice_can);
}