import type { Claims } from "../../types/claims";
import type {Signal, Sdp, Ice, PeerEvent, PcType, SdpType, ActionType, PeerAction} from "../../types/signal";
import Denque from "denque"

export class SignalClient {
    private signal_url: string;
    private ws!: WebSocket;
    private queue = new Denque<Signal>();

    private send(msg: Signal) {
        if (this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(msg));
        }else {
            this.queue.push(msg);
        }
    }

    private _onClose?: (ev: CloseEvent) => void;
    private _onError?: (ev: Event) => void;
    private _onSdp?: (payload: Sdp) => void;
    private _onIce?: (payload: Ice) => void;

    // Allow multiple event callbacks
    private _onEventSubs = new Set<(e: PeerEvent) => void>()
    
    
    constructor(ws_url: string) {
        this.signal_url = ws_url;
    }

    connect(){
        this.ws = new WebSocket(this.signal_url);
        this.ws.onopen = () => {
            while (this.queue.length) {
                this.ws.send(JSON.stringify(this.queue.shift()!));
            }
        };

        this.ws.onclose = (ev) => this._onClose?.(ev);
        this.ws.onerror = (ev) => this._onError?.(ev);

        this.ws.onmessage = (ev) => {
            const msg = JSON.parse(ev.data) as Signal;
            switch (msg.type) {
                case "sdp": this._onSdp?.(msg.payload); break;
                case "ice": this._onIce?.(msg.payload); break;
                case "event": 
                    for (const cb of this._onEventSubs) {
                        try {cb(msg.payload);}catch(e) {console.error(e)}
                    }
                    break;

                default:
            }
        }
    }

    close(code: number = 1000, reason: string= "client shutdown"): void {
        
        if (this.ws.readyState === WebSocket.CLOSED) {
            this.cleanup()
            return;
        } else if (this.ws.readyState === WebSocket.CONNECTING || this.ws.readyState === WebSocket.OPEN) {
            this.ws.close(code, reason);
            this.cleanup()
        }
    }
    
    // atttach functions
    onClose(fn: (ev: CloseEvent) => void) {this._onClose = fn};
    onError(fn: (ev: Event) => void) {this._onError = fn};
    onSdp(fn: (sdp: Sdp) => void) {this._onSdp = fn};    
    onIce(fn: (ice: Ice) => void) {this._onIce = fn};
    onEvent(fn: (e: PeerEvent) => void) {this._onEventSubs.add(fn); return () => this._onEventSubs.delete(fn)};


    sendSdp(pc: PcType, type: SdpType, sdp: string) {
        this.send({type: "sdp", payload: {pc, type, sdp}})
    }

    sendIce(pc: PcType, ice: RTCIceCandidate) {
        const payload: Ice = {
            pc,
            candidate: ice.candidate,
            sdpMid: ice.sdpMid ?? "",
            sdpMLineIndex: ice.sdpMLineIndex ?? 0,
            usernameFragment: ice.usernameFragment ?? "",
        };

        this.send({type: "ice", payload})
    }

    sendAction(claims: Claims, action: ActionType) {
        const payload: PeerAction = {
            peerID: claims.ID,
            roomID: claims.roomID,
            type: action,
            role: claims.role,
        };

        this.send({type: "action", payload});
    }

    private cleanup() {
        this.ws.onopen = null;
        this.ws.onmessage = null;
        this.ws.onclose = null;
        this.ws.onerror = null;

        while (this.queue.length) this.queue.shift();

    }

}