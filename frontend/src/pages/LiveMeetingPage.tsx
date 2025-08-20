import { useEffect, useRef } from "react";
import { sub_conn, RtcClient } from "../services/lib/webrtc";

export default function MeetingPage() {
  const clientRef = useRef<RtcClient | null>(null);
  const v0 = useRef<HTMLVideoElement>(null);
  const v1 = useRef<HTMLVideoElement>(null);
  const v2 = useRef<HTMLVideoElement>(null);

  useEffect(() => {
    clientRef.current = sub_conn;

    // Poll a few times per second to (re)bind streams
    const id = setInterval(() => {
      const streams = sub_conn.remoteVideos;
      const els = [v0.current, v1.current, v2.current];

      for (let i = 0; i < 3; i++) {
        const el = els[i];
        const s = streams[i]; // MediaStream | null
        if (!el) continue;
        if (el.srcObject !== s) {
          el.srcObject = s as any;
          // Kick playback (helps on some browsers)
          el.muted = true;       // required for autoplay
          el.playsInline = true; // avoid iOS fullscreen
          el.play().catch(() => {});
        }
      }
    }, 300);

    return () => {
      clearInterval(id);
      sub_conn.pc.close();
      clientRef.current = null;
    };
  }, []);

  return (
    <div style={{ display: "grid", gap: 12, gridTemplateColumns: "repeat(3, 1fr)", padding: 12 }}>
      <video ref={v0} autoPlay playsInline muted style={{ width: "100%", height: 200, objectFit: "cover", background: "black" }} />
      <video ref={v1} autoPlay playsInline muted style={{ width: "100%", height: 200, objectFit: "cover", background: "black" }} />
      <video ref={v2} autoPlay playsInline muted style={{ width: "100%", height: 200, objectFit: "cover", background: "black" }} />
    </div>
  );
}