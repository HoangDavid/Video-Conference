import { useEffect, useRef } from "react";
import { sub_conn, waitUntilDualPcConnect } from "../services/lib/webrtc";
export default function MeetingPage() {
    const remoteRef = useRef<HTMLVideoElement>(null)
    const remoteRef1 = useRef<HTMLVideoElement>(null)

    useEffect(() => {
      (async() => {
        if (remoteRef.current && remoteRef1.current && await waitUntilDualPcConnect()) {
          remoteRef.current!.srcObject = sub_conn.remoteStream
          remoteRef.current.muted = true 
          remoteRef.current?.play().catch((e) => {console.log(e)})

          remoteRef1.current!.srcObject = sub_conn.remoteStream1
          remoteRef1.current.muted = true 
          remoteRef1.current?.play().catch((e) => {console.log(e)})

          console.log(remoteRef)
        }
      })()
    })




  return (
    <>
    <video ref={remoteRef} style={{ width: "50vw" }} playsInline />
    <video ref={remoteRef1} style={{ width: "50vw" }} playsInline />
    </>
  );
}