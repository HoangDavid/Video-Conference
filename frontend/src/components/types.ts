
export interface Signal{
    type: 'offer' | "answer" | "ice";
    payload: unknown;
}

export interface Ice {
    candidate: string
    sdpMid?: string
    sdpMLineIndex?: number
}