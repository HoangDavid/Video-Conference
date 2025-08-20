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
    
    
    constructor(ws_url: string) {this.signal_url = ws_url;}

    connect(){
        if (this.ws && (this.ws.readyState === WebSocket.OPEN || this.ws.readyState === WebSocket.CONNECTING)) return;

        const ws = new WebSocket(this.signal_url);
        this.ws = ws;

        ws.onopen = () => {
            while (this.queue.length) ws.send(JSON.stringify(this.queue.shift()!));
        };
        ws.onclose = (ev) => this._onClose?.(ev);
        ws.onerror = (ev) => this._onError?.(ev);
        ws.onmessage = (ev) => {
            const msg = JSON.parse(ev.data) as Signal;
            switch (msg.type) {
                case "sdp": this._onSdp?.(msg.payload); break;
                case "ice": this._onIce?.(msg.payload); break;
                case "event":
                for (const cb of this._onEventSubs) { try { cb(msg.payload); } catch (e) { console.error(e); } }
                break;
            }
        };
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

    sendAction(action: ActionType) {
        const payload: PeerAction = {
            type: action,
        };

        this.send({type: "action", payload});
    }

    // true  if success, false otherwise
    isOpen(): boolean { return !!this.ws && this.ws.readyState === WebSocket.OPEN; }

    async waitUnillOpen(timeoutMs = 5000): Promise<boolean> {
        if (this.isOpen()) return true;
            this.connect();
            const ws = this.ws!;
        return await new Promise<boolean>((resolve) => {
            const done = (ok: boolean) => {
                ws.removeEventListener("open", onOpen);
                ws.removeEventListener("error", onErr);
                clearTimeout(t);
                resolve(ok);
            };
            const onOpen = () => done(true);
            const onErr = () => done(false);
            const t = setTimeout(() => done(false), timeoutMs);
            ws.addEventListener("open", onOpen, { once: true });
            ws.addEventListener("error", onErr, { once: true });
        });
    }

    private cleanup() {
        this.ws.onopen = null;
        this.ws.onmessage = null;
        this.ws.onclose = null;
        this.ws.onerror = null;

        while (this.queue.length) this.queue.shift();

    }

}

export const signal_conn = new SignalClient("/ws")