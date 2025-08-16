import { SignalClient } from "./signal"; 
import type { Sdp, Ice } from "../../types/signal";

export class RtcClient {
    readonly pc: RTCPeerConnection;
    readonly remoteStream = new MediaStream()
    AVattached: boolean = false;

    private pendingIce: RTCIceCandidateInit[] = []

    private _onIce?: (ice: RTCIceCandidate) => void;
    private _onTrack?: (track: MediaStreamTrack) => void;
    onConnectionStateChange?: (state: RTCPeerConnectionState) => void;
    
    constructor(){
        this.pc =  new RTCPeerConnection({
            iceServers: [{ urls: "stun:stun.l.google.com:19302" }],
        });

        this.pc.onicecandidate = (e) => {
            if (e.candidate) this._onIce?.(e.candidate)
            else return;
        };

        this.pc.ontrack = (ev) => {
            // TODO: figure this out
        };
    };

    onIce(fn: (ice: RTCIceCandidate) => void) {this._onIce = fn};
    
    // attach local medias
    attachLocalStream(stream: MediaStream) {
        if (!this.AVattached) {
            stream.getTracks().forEach((t) => this.pc.addTrack(t, stream));
            this.AVattached = true;
        }
        
    }

    // create offer and set local description
    async createOfferAndSetLocal(): Promise<string>{
        const offer = await this.pc.createOffer();
        await this.pc.setLocalDescription(offer);
        console.log(this.pc.localDescription?.sdp)
        return this.pc.localDescription?.sdp ?? "";
    }

    // set remote answer
    async setRemoteAnswer(sdp: string) {
        await this.pc.setRemoteDescription({type: "answer", sdp});
        await this.flushBufferedIce();
    }

    // create answer and set local description
    async answerRemoteOffer(sdp: string): Promise<string> {
        await this.pc.setRemoteDescription({type: "offer", sdp})
        const answer = await this.pc.createAnswer();
        await this.pc.setLocalDescription(answer);
        await this.flushBufferedIce();
        return answer.sdp ?? "";
    }

    //  add ice candidate
    async setRemoteIce(ice: RTCIceCandidateInit) {
        if (this.pc.remoteDescription) await this.pc.addIceCandidate(ice);
        else this.pendingIce.push(ice);
    } 

    async waitForPc(timeoutMs = 5000): Promise<boolean> {
        return new Promise<boolean>((resolve) => {
            if (this.pc.connectionState === "connected") return resolve(true);

            const onChange = () => {
                const s = this.pc.connectionState;
                if (s === "connected") { cleanup(this.pc); resolve(true); }
                else if (s === "failed" || s === "disconnected" || s === "closed") { cleanup(this.pc); resolve(false); }
            };

            const t = setTimeout(() => { cleanup(this.pc); resolve(this.pc.connectionState === "connected"); }, timeoutMs);

            this.pc.addEventListener("connectionstatechange", onChange);
            function cleanup(pc: RTCPeerConnection) {
                pc.removeEventListener("connectionstatechange", onChange);
                clearTimeout(t);
            }
        });
    }

    //  flush ice candidate
    private async flushBufferedIce() {
        for (const c of this.pendingIce) await this.pc.addIceCandidate(c);
        this.pendingIce.length = 0;
    }

}

// Initialize two peer connection
export const pub_conn = new RtcClient();
export const sub_conn = new RtcClient();

export function wireCallBacks(conn: SignalClient){
    conn.onSdp((sdp: Sdp) => {
        if (sdp.pc == "pub") {
            HandleRemoteSdp(sdp, pub_conn, conn);
        }else if (sdp.pc == "sub") {
            HandleRemoteSdp(sdp,sub_conn, conn);
        }
    })
    
    conn.onIce((ice: Ice) => {
        if (ice.pc == "pub") {
            HandleRemoteIce(ice, pub_conn);
        }else if (ice.pc == "sub") {
            HandleRemoteIce(ice, sub_conn);
        }
    })
    
    //  Attach ice  callbacks to peer conn
    pub_conn.onIce((ice: RTCIceCandidate) => {
        conn.sendIce("pub", ice);
    })
    
    sub_conn.onIce((ice: RTCIceCandidate) => {
        conn.sendIce("sub", ice);
    })
}


// wait until both pc connects
export async function waitUntilDualPcConnect(conn: SignalClient): Promise<boolean>{
    if (!pub_conn.AVattached) {
        console.error("publisher av is not attached!");
        return false
    }
    const offer = await pub_conn.createOfferAndSetLocal();

    conn.sendSdp("pub","offer", offer);

    if (!pub_conn.waitForPc() || !sub_conn.waitForPc()) {
        return false
    }
    return true
}


// helper function: handle sdp from remote peer
async function HandleRemoteSdp(sdp: Sdp, pc_conn: RtcClient, signal_conn: SignalClient) {
    if (sdp.type == "answer") {
        await  pc_conn.setRemoteAnswer(sdp.sdp);
    }else if (sdp.type == "offer") {
        const answer = await pc_conn.answerRemoteOffer(sdp.sdp);
        if (!answer) {return}
        signal_conn.sendSdp(sdp.pc, "answer", answer)
    }
}


// helper function: handle Ice from remote peer
function HandleRemoteIce(ice: Ice, p_conn: RtcClient) {
    const ice_can: RTCIceCandidateInit = {
        candidate: ice.candidate,
        sdpMid: ice.sdpMid,
        sdpMLineIndex: ice.sdpMLineIndex,
    };

    p_conn.setRemoteIce(ice_can);
}