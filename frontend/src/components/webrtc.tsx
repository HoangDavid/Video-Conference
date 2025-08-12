import { useEffect, useRef} from "react";
import type { Signal, Ice} from "./webrtc";



export default function OnePeerClient() {
  const pc = useRef<RTCPeerConnection | null>(null);
  const ws = useRef<WebSocket | null>(null);

  const localRef  = useRef<HTMLVideoElement>(null);
  const remoteRef = useRef<HTMLVideoElement>(null);

  const pending: RTCIceCandidateInit[] = [];

  function buffer(c: RTCIceCandidateInit) {
    pending.push(c);          // producer
  }

  async function flush(pc: RTCPeerConnection) {
    for (const c of pending) await pc.addIceCandidate(c);
    pending.length = 0;       // consumer
  }


  useEffect(() => {
    (async () => {
      /* 1. local media preview */
      const stream = await navigator.mediaDevices.getUserMedia({ video: true, audio: true});
      
      if (localRef.current) {
        localRef.current.srcObject = stream;
        localRef.current.muted = true;
        await localRef.current.play();
      }

      /* 2. WS + PeerConnection */
      ws.current = new WebSocket("wss://localhost:8443/ws");
      pc.current = new RTCPeerConnection({
        iceServers: [{ urls: "stun:stun.l.google.com:19302" }],
      });

      stream.getTracks().forEach(t => pc.current!.addTrack(t, stream));
      const remoteStream = new MediaStream();
      if (remoteRef.current) remoteRef.current!.srcObject = remoteStream;

      pc.current.ontrack = ev => {
        console.log("new remote track", ev.track.kind, ev.track.id);
        remoteStream.addTrack(ev.track);
        remoteRef.current?.play()
      }

      pc.current.onicecandidate = e => {
        if (!e.candidate) return;
        const msg: Signal = {
          type: "ice",
          payload: e.candidate.toJSON() as Ice,
        };
        ws.current!.send(JSON.stringify(msg));
        console.log("ice sent")
      };      

      ws.current.onmessage = async ev => {
        const msg = JSON.parse(ev.data) as Signal;

        if (msg.type === "answer") {
          const { sdp } = msg.payload as { sdp: string };
          await pc.current!.setRemoteDescription(
            new RTCSessionDescription({ type: "answer", sdp })
          );
          console.log("remote description set")
          flush(pc.current!)
          console.log("add ice")
          
        }

        // TODO: add recieving offer to do renegotiation

        if (msg.type === "ice") {
          // console.log(msg.payload)
          const payload  = msg.payload as Ice
          const ice: RTCIceCandidateInit = {
            candidate: payload.candidate,
            sdpMid: payload.sdpMid,
            sdpMLineIndex: payload.sdpMLineIndex
          }

          if (pc.current!.remoteDescription) {
            await pc.current!.addIceCandidate(ice)
            console.log("add ice")
          }else {
            buffer(ice)
            console.log("buffer ice")
          }
            

        }
      };


      pc.current!.onconnectionstatechange = () =>
        console.log("PC state →", pc.current!.connectionState)


      ws.current.onopen = async () => {
        const offer = await pc.current!.createOffer();
        await pc.current!.setLocalDescription(offer);
        
        const msg: Signal = {
          type: "offer",
          payload: { sdp: offer.sdp },
        };

        // send offer
        ws.current!.send(JSON.stringify(msg));
        console.log("offer sent")
      };

    })();

    return () => {
      pc.current?.close();
      ws.current?.close();
    };
  }, []);

  

  return (
    <>
      <video ref={localRef}  style={{ width: 300 }} autoPlay playsInline />
      <video ref={remoteRef} style={{ width: 300 }} playsInline />
    </>
  );
}