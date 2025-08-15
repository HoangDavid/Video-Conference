import CircularProgress from '@mui/material/CircularProgress';
import MicOffIcon from '@mui/icons-material/MicOff';
import MicIcon from '@mui/icons-material/Mic';
import VideocamIcon from '@mui/icons-material/Videocam';
import VideocamOffIcon from '@mui/icons-material/VideocamOff';
import { styled } from '@mui/system';
import { useState, useRef, useEffect } from 'react';
import { Toaster, toast } from "react-hot-toast";
import { getSignal } from '../services/lib/instance';
import type { PeerEvent } from '../types/signal';
import { useNavigate } from "react-router-dom"; 
import '../App.css';


export default function Lobby(){

    // Logic:
    const navigate = useNavigate();
    const signal_conn = getSignal();
    const mediaRef = useRef<HTMLVideoElement>(null);
    const [audio, setAudio] = useState<boolean>(true);
    const [video, setVideo] = useState<boolean>(true);
    
    // Callback to reroute to meeting if active
    if (signal_conn) {
        const del = signal_conn?.onEvent((e: PeerEvent) => {
            if (e.type == "room_active") {
                navigate(`/rooms/${e.roomID}/meeting`);
                del();
            }
        })
    }

    // Video preview
    useEffect(() => {
        let stream: MediaStream;
        (async() => {
            stream = await navigator.mediaDevices.getUserMedia({video: true, audio: true});
            
            if(mediaRef.current) {
                mediaRef.current.srcObject = stream;
                await mediaRef.current.play().catch(() => {});
            }
        })();

        return () => {
            stream?.getTracks().forEach(t => t.stop())
        };
    }, []);

    const toggleVideo = (flag: boolean) => {
        if (mediaRef.current && mediaRef.current.srcObject as MediaStream) {
            const s = mediaRef.current.srcObject as MediaStream;
            s?.getVideoTracks().forEach(t => (t.enabled = flag));
            setVideo(flag)
        }
    }

    const toggleAudio = (flag: boolean) => {
        if (mediaRef.current && mediaRef.current.srcObject as MediaStream) {
            const s = mediaRef.current.srcObject as MediaStream;
            s?.getAudioTracks().forEach(t => (t.enabled = flag));
            setAudio(flag)
        }
    }


    // Customization
    const videoFrame: React.CSSProperties = {
        width: "350px",
        position: "relative"
    }

    const videoBox: React.CSSProperties = {
        width: "350px",
        borderRadius: "10px",
    }

    const mediaBar: React.CSSProperties = {
        background: "transparent",
        color: "lightgray",
        padding: "10px",
        position: "absolute",
        left: "50%",
        bottom: "10px",            // add px
        transform: "translateX(-50%)",
        display: "flex",
        alignItems: "center",
        gap: "16px",

    }

    const mediaOnBtn = {
        border: "2px solid",
        borderRadius: "50%",
        cursor: "pointer",
        background: "transparent",
        color: "lightgray",
        padding: "10px",
    }
    
    const mediaOffBtn = { 
        border: "2px solid",
        borderRadius: "50%",
        cursor: "pointer",
        background: "#D10000",
        color: "lightgray",
        padding: "10px",
    }

    return <>
    <h1> Lobby </h1>
    <div>
        <div style={videoFrame}>
            <video ref={mediaRef} autoPlay playsInline style={videoBox}/>
            <div style={mediaBar}>
                {audio && <MicIcon style={mediaOnBtn} onClick={() => toggleAudio(!audio)}/>}
                {!audio && <MicOffIcon style={mediaOffBtn} onClick={() => toggleAudio(!audio)}/>}

                {video && <VideocamIcon style={mediaOnBtn}  onClick={() => toggleVideo(!video)}/>}
                {!video && <VideocamOffIcon style={mediaOffBtn}  onClick={() => toggleVideo(!video)}/>}
            </div>

        </div>
        
        <div style={{marginTop: "10px"}}>
            <CircularProgress/>
            <p> Wait for your host to start the meeting</p>
        </div>
    </div>
    </>
}