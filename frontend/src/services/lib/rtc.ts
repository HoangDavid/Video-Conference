

export class RtcClient {
    readonly pc: RTCPeerConnection;
    readonly remoteStream = new MediaStream()

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
        stream.getTracks().forEach((t) => this.pc.addTrack(t, stream));
    }

    // create offer and set local description
    async createOfferAndSetLocal(): Promise<string>{
        const offer = await this.pc.createOffer();
        await this.pc.setLocalDescription(offer);
        return offer.sdp ?? "";
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

    //  flush ice candidate
    private async flushBufferedIce() {
        for (const c of this.pendingIce) await this.pc.addIceCandidate(c);
        this.pendingIce.length = 0;
    }

}