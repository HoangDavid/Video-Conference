import { SignalClient } from "./signal"; 
import type { Sdp, Ice } from "../../types/signal";

export class RtcClient {
    private pc!: RTCPeerConnection;
    public remoteStream: MediaStream =  new MediaStream();


    AVattached: boolean = false;


    private pendingIce: RTCIceCandidateInit[] = []

    private _onIce?: (ice: RTCIceCandidate) => void;
    private _onTrack?: (track: MediaStreamTrack, stream?: MediaStream) => void;
    onConnectionStateChange?: (state: RTCPeerConnectionState) => void;
    

    constructor() {}
    connect(){
        this.pc =  new RTCPeerConnection({
            iceServers: [{ urls: "stun:stun.l.google.com:19302" }],
        });

        this.pc.onicecandidate = (e) => {
            if (e.candidate) this._onIce?.(e.candidate)
            else return;
        };

        this.pc.ontrack = (ev) => {
            this.remoteStream?.addTrack(ev.track)
        };

        setInterval(async () => {
            const stats = await this.pc.getStats();

            stats.forEach(r => {
                if (r.type === "outbound-rtp" && r.kind === "video") {
                    console.log(`[stats] video framesSent=${r.framesSent}`);
                
                }
                if (r.type === "inbound-rtp" && r.kind === "video") {
                    console.log(`[recv]  video pli count  = ${r.pliCount}`);
                    console.log(`[recv]  video framesDecoded  = ${r.framesDecoded}`);
                    console.log(`[recv]  video packetsLost    = ${r.packetsLost}`);
                }

             });
         }, 2000);
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
                else if (s === "failed" || s === "disconnected" || s === "closed") { cleanup(this.pc); console.log("connection state: ", s); resolve(false); }
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


export function wireCallBacks(pub_conn: RtcClient, sub_conn: RtcClient, conn: SignalClient){
    conn.onSdp((sdp: Sdp) => {
        console.log("got sdp type ", sdp.type, sdp.pc)
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

export const pub_conn = new RtcClient()
export const sub_conn = new RtcClient()

export async function pcConnect( conn: SignalClient, stream: MediaStream): Promise<boolean> {
    pub_conn.connect()
    sub_conn.connect()

    pub_conn.attachLocalStream(stream)
    wireCallBacks(pub_conn, sub_conn, conn)

    const offer = await pub_conn.createOfferAndSetLocal();
    conn.sendSdp("pub","offer", offer);


    const conn1 =  await pub_conn.waitForPc()
    const conn2 = await sub_conn.waitForPc()
    console.log("pub", conn1,"sub", conn2)

    if (!conn1|| !conn2) {
        return false
    }

    return true

}


// wait until both pc connects
export async function waitUntilDualPcConnect(): Promise<boolean>{
    if (!pub_conn.AVattached) {
        console.error("publisher av is not attached!");
        return false
    }

    const conn1 =  await pub_conn.waitForPc()
    const conn2 = await sub_conn.waitForPc()
    console.log("pub", conn1,"sub", conn2)
    
    if (!conn1|| !conn2) {
        return false
    }

    return true
}


// helper function: handle sdp from remote peer
async function HandleRemoteSdp(sdp: Sdp, pc_conn: RtcClient, signal_conn: SignalClient) {
    if (sdp.type == "answer") {
        await  pc_conn.setRemoteAnswer(sdp.sdp);
        console.log("set remote answer sdp for: ", sdp.pc, sdp.type)
        
    }else if (sdp.type == "offer") {
        const answer = await pc_conn.answerRemoteOffer(sdp.sdp);
        if (!answer) {return}
        signal_conn.sendSdp(sdp.pc, "answer", answer)
        console.log("send answer to: ", sdp.pc)
    }
}


// helper function: handle Ice from remote peer
function HandleRemoteIce(ice: Ice, p_conn: RtcClient) {
    console.log("got ice from, ", ice.pc)
    const ice_can: RTCIceCandidateInit = {
        candidate: ice.candidate,
        sdpMid: ice.sdpMid,
        sdpMLineIndex: ice.sdpMLineIndex,
    };

    p_conn.setRemoteIce(ice_can);
}
