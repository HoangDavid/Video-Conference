import { useEffect, useRef} from "react";
import type { Signal} from "./types";

export default function OnePeerClient() {
    const pc = useRef<RTCPeerConnection | null>(null)
    const ws = useRef<WebSocket | null>(null)

    const localRef  = useRef<HTMLVideoElement>(null)
    const remoteRef = useRef<HTMLVideoElement>(null)

    useEffect(() => {
        (async() => {
            const stream = await navigator.mediaDevices.getUserMedia({audio: false, video: true})
            if (localRef.current){
                localRef.current.srcObject = stream
                localRef.current.muted = true
                await localRef.current.play()
            }

            ws.current = new WebSocket("wss://localhost:8443/ws")
            pc.current = new RTCPeerConnection({
                iceServers: [{ urls: 'stun:stun.l.google.com:19302' }],
            })

            stream.getTracks().forEach(t => pc.current!.addTrack(t, stream))

            pc.current.onicecandidate = e => {
                if (e.candidate) {
                    const msg: Signal = {
                        type: "ice",
                        payload: {candidate: e.candidate.toJSON()}
                    }
                     ws.current!.send(JSON.stringify(msg))
                }
            }


            pc.current.ontrack = e => {
                console.log('🔔 ontrack kind=', e.track.kind, 'id=', e.track.id);
                // one MediaStream per video element
                const rs = (remoteRef.current!.srcObject ||
                new MediaStream()) as MediaStream

                if (!rs.getTracks().some(t => t.id === e.track.id)) {
                    rs.addTrack(e.track)
                    remoteRef.current!.srcObject = rs
                }
            }

            
            const pending: RTCIceCandidateInit[] = [];
            ws.current.onmessage = async e => {
                const msg = JSON.parse(e.data) as Signal;

                if (msg.type === "answer") {
                    const { sdp } = msg.payload as { sdp: string };
                    console.log("setting remote")
                    await pc.current!.setRemoteDescription(
                    new RTCSessionDescription({ type: "answer", sdp })
                    );

                    /* flush any ICE that arrived early */
                    for (const ice of pending) {
                    await pc.current!.addIceCandidate(ice);
                    }
                    pending.length = 0;                  // clear queue
                }

                if (msg.type === "ice") {
                    const ice = typeof msg.payload === "string"
                    ? JSON.parse(msg.payload)
                    : msg.payload;
                    
                    if (!('sdpMid' in ice) && !('sdpMLineIndex' in ice)) {
                        ice.sdpMid        = '0';  // or the mid string you see in the SDP ("video")
                        ice.sdpMLineIndex = 0;    // first m-line
                    }

                    console.log('⬅  ICE from server:', ice);

                    if (pc.current!.remoteDescription) {
                        console.log("adding Ice")
                        await pc.current!.addIceCandidate(ice as RTCIceCandidateInit);
                    } else {
                        pending.push(ice as RTCIceCandidateInit);
                    }
                }
                
            }

            ws.current.onopen = async() => {
                const offer = await pc.current!.createOffer()
                await pc.current!.setLocalDescription(offer)
                ws.current!.send(
                    JSON.stringify({type: "offer", payload: {sdp: offer.sdp}})
                )
            }

        })()
        
    return () => {
      pc.current?.close()
      ws.current?.close()
    }
    }, [])

    return <>
      <video ref={localRef} style={{ width: 300 }} />
      <video ref={remoteRef} style={{ width: 300 }} autoPlay playsInline />
    </>
}