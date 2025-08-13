import type {Signal, Sdp, Ice, PeerAction, PeerEvent, PcType, SdpType} from "../../types/signal";
import Denque from "denque"

export class SignalClient {
    private ws: WebSocket;
    private queue = new Denque<Signal>();

    private send(msg: Signal) {
        if (this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(msg));
        }else {
            this.queue.push(msg);
        }
    }

    onClose?: (ev: CloseEvent) => void;
    onError?: (ev: Event) => void;

    onSdp?: (payload: Sdp) => void;
    onIce?: (payload: Ice) => void;
    onEvent?: (payload: PeerEvent) => void;
    
    
    constructor(ws_url: string) {
        this.ws = new WebSocket(ws_url);
        this.ws.onopen = () => {
            while (this.queue.length) {
                this.ws.send(JSON.stringify(this.queue.shift()!));
            }
        };

        this.ws.onclose = (ev) => this.onClose?.(ev);
        this.ws.onerror = (ev) => this.onError?.(ev);

        this.ws.onmessage = (ev) => {
            const msg = JSON.parse(ev.data) as Signal;
            switch (msg.type) {
                case "sdp": this.onSdp?.(msg.payload); break;
                case "ice": this.onIce?.(msg.payload); break;
                case "event": this.onEvent?.(msg.payload); break;
                default:
            }
        }
    }

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

    sendAction(action: PeerAction) {
        this.send({type: "action", payload:action})
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

    private cleanup() {
        this.ws.onopen = null;
        this.ws.onmessage = null;
        this.ws.onclose = null;
        this.ws.onerror = null;

        while (this.queue.length) this.queue.shift();

    }

}