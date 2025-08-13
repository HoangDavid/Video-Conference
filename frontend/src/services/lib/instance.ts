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
let rtc: RtcClient | null = null;

export function initRtc() {
    if (!rtc) rtc = new RtcClient();
    return rtc;
}



