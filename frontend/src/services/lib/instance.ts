import { SignalClient } from "./signal";
import { RtcClient } from "./rtc";

// signaling singleton
let signal: SignalClient | null = null;

export function initSignal(ws_url: string){
    if (!signal) signal = new SignalClient(ws_url);
    return signal
}

export function getSignal(): SignalClient | null {
    if (!signal) { 
        console.error("signal client is not initialized");
        return null
    }
    return signal;
}

export function destroySignal() {
    signal?.close();
    signal = null;
}

// webrtc connetion singleton
let pub: RtcClient | null = null;
let sub: RtcClient | null = null;

export function initRtcPub() {
    if (!pub) pub = new RtcClient();
    return pub;
}

export function initRtcSub() {
    if (!sub) sub = new RtcClient();
    return sub
}

export function getRtcPub (){
    if (!pub) { 
        console.error("publisher is not initialized");
        return null
    }

    return pub
}

export function getRtcSub () {
    if (!sub) {
        console.error("subcriber is not initialized");
        return null
    }

    return sub
}



