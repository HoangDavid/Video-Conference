import { useEffect, useRef, useState } from "react";
import MicOffIcon from '@mui/icons-material/MicOff';
import MicIcon from '@mui/icons-material/Mic';
import VideocamIcon from '@mui/icons-material/Videocam';
import VideocamOffIcon from '@mui/icons-material/VideocamOff';
import SubtitlesIcon from '@mui/icons-material/Subtitles';
import SubtitlesOffIcon from '@mui/icons-material/SubtitlesOff';
import LogoutIcon from '@mui/icons-material/Logout';
import HistoryIcon from '@mui/icons-material/History';
import TranslateIcon from '@mui/icons-material/Translate';

import { sub_conn, waitUntilDualPcConnect } from "../services/lib/webrtc";
import { media } from "../services/lib/media";



export default function MeetingPage() {

    // Logic
    const localRef = useRef<HTMLVideoElement>(null)
    const remoteRef = useRef<HTMLVideoElement>(null)
    const [audio, setAudio] = useState(false)
    const [video, setVideo] = useState(false)
    const [sub, setSub] = useState(false)
    const [dub, setDub]= useState(false)
    const [loading, setLoading] = useState(false)

    useEffect(() => {
      (async() => {
        setLoading(true)
        if (remoteRef.current && await waitUntilDualPcConnect()) {
          remoteRef.current!.srcObject = sub_conn.remoteStream
          remoteRef.current.muted = false 
          remoteRef.current?.play().catch((e) => {console.log(e)})
        }

        const m = await media.getAV()
        if (localRef.current){
          localRef.current!.srcObject = m
          localRef.current.muted = true;
          await localRef.current.play().catch((err) => {console.log(err)})
        }

        if (media.audio?.enabled) setAudio(true);
        if (media.video?.enabled) setVideo(true)
        setLoading(false)
      })()
    })

    const toggleAudio = () => {
      if (loading) return;
      
      setLoading(true)
      media.set_Mic(!(media.audio?.enabled ?? false));
      setAudio(media.audio?.enabled ?? false);
      setLoading(false)
    }

    const toggleVideo = () => {
      if (loading) return;

      setLoading(true)
      media.set_Video(!(media.video?.enabled ?? false));
      setVideo(media.video?.enabled ??  false);
      setLoading(false)
    }

    const toggleSub = () => {
      if (loading) return;

      setLoading(true)
      setSub(!sub)
      setLoading(false)
    }

    const toggleDub = () => {
      if (loading) return;

      setLoading(true)
      setDub(!dub)
      setLoading(false)
    }

    // Customization
    const remoteFrame: React.CSSProperties = {
      width: "50vw",
      left: 0,
      right: 0,
      margin: "auto",
      position: "relative",
    }

    const remoteVideo: React.CSSProperties = {
      width: "inherit",
      borderRadius: "15px",
      left: 0,
      right: 0,
      margin: "auto"
    }


    const localFrame: React.CSSProperties = {
      width: "15vw",
      position: "absolute",
      top: 20,
      left: 20
    }

    const nameLabel: React.CSSProperties = {
      position: "absolute",
      top: 0,
      left: 0,
      backgroundColor: "rgba(0,0,0,0.5)",
      borderRadius: "15px 0px 0px 0px",
      padding: "5px"
    }

    const localVideo: React.CSSProperties = {
      width: "inherit",
      borderRadius: "15px",
    }

    const utilityBar: React.CSSProperties = {
      width: "100%",
      borderTop: "grey solid",
      borderBottom: "grey solid",
      position: "absolute",
      bottom: 0,
      left: 0,
      height:"100px",
      display: "flex",
      alignItems: "center",
      gap: "16px",
      justifyContent: "center",
    }

    const mediaOnBtn: React.CSSProperties = {
        border: "2px solid",
        borderRadius: "50%",
        cursor: "pointer",
        background: "transparent",
        color: "lightgray",
        padding: "10px",
        position: "relative",
    }
    
    const mediaOffBtn: React.CSSProperties = { 
        border: "2px solid",
        borderRadius: "50%",
        cursor: "pointer",
        background: "#D10000",
        color: "lightgray",
        padding: "10px",
        position: "relative",
    }

    


  return (
    <>
      <div style={remoteFrame}>
        <div style={nameLabel}> Remote </div>
        <video ref={remoteRef} style={remoteVideo} playsInline />
      </div>

      <div style={localFrame}>
        <div style={nameLabel}>
          You
        </div>
        <video ref={localRef} style={localVideo} playsInline/>
      </div>

      <div style={utilityBar}>
        {!audio && <MicOffIcon  onClick={() => toggleAudio()} style={mediaOffBtn}/>}
        {audio && <MicIcon  onClick={() => toggleAudio()} style={mediaOnBtn}/>}

        {!video && <VideocamOffIcon  onClick={() => toggleVideo()} style={mediaOffBtn}/>}
        {video && <VideocamIcon   onClick={() => toggleVideo()} style={mediaOnBtn}/>}

        {!sub && <SubtitlesOffIcon onClick={() => toggleSub()} style={mediaOnBtn}/>}
        {sub && <SubtitlesIcon onClick={() => toggleSub()} style={mediaOnBtn}/>}

        <TranslateIcon style={mediaOnBtn}/>
        <LogoutIcon style={mediaOffBtn}/>
      </div>
    </>
  );
}