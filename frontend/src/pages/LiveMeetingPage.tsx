import { useEffect, useRef } from "react";
import { sub_conn, waitUntilDualPcConnect } from "../services/lib/webrtc";
export default function MeetingPage() {
    const remoteRef = useRef<HTMLVideoElement>(null)

    useEffect(() => {
      (async() => {
        if (remoteRef.current && await waitUntilDualPcConnect()) {
          remoteRef.current!.srcObject = sub_conn.remoteStream
          remoteRef.current.muted = true 
          remoteRef.current?.play().catch((e) => {console.log(e)})
          console.log(remoteRef)
        }
      })()
    })




  return (
    <video ref={remoteRef} style={{ width: "50vw" }} playsInline />
  );
}