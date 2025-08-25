import { useRef, useState, useEffect } from 'react';
import { media } from '../services/lib/media';

import MicOffIcon from '@mui/icons-material/MicOff';
import MicIcon from '@mui/icons-material/Mic';
import VideocamIcon from '@mui/icons-material/Videocam';
import VideocamOffIcon from '@mui/icons-material/VideocamOff';
import CircularProgress from '@mui/material/CircularProgress';

export default function PreviewVideo(){
    // Logic
    const mediaRef = useRef<HTMLVideoElement>(null);
    const [audio, setAudio] = useState<boolean>(true);
    const [video, setVideo] = useState<boolean>(true);
    const [loading, setLoading] = useState<boolean>(true)

    useEffect(() => {
        (async() => {
            const m = await media.getAV();
            if (mediaRef.current){
                mediaRef.current.srcObject = m;
                mediaRef.current.muted = true;
                await mediaRef.current.play().catch((e) => {console.error(e)})
            }
            setAudio(media.audio?.enabled ?? false);
            setVideo(media.video?.enabled ?? false);
            setLoading(false);
            
        })()

    }, [])

    const toggleAudio = () => {
        media.set_Mic(!(media.audio?.enabled ?? false));
        setAudio(media.audio?.enabled ?? false);
    }

    const toggleVideo = () => {
        media.set_Video(!(media.video?.enabled ?? false));
        setVideo(media.video?.enabled ??  false);
    }

    // Cusotomization
    const videoFrame: React.CSSProperties = {
        width: "40vw",
        position: "relative",
        verticalAlign: "middle",
        left: 0,
        right: 0,
        margin: "auto"
    }

    const videoBox: React.CSSProperties = {
        width: "40vw",
        borderRadius: "10px",
    }

    const mediaBar: React.CSSProperties = {
        background: "transparent",
        color: "lightgray",
        padding: "10px",
        position: "absolute",
        left: "50%",
        bottom: "10px",
        transform: "translateX(-50%)",
        display: "flex",
        alignItems: "center",
        gap: "16px",
    }

    const mediaOnBtn: React.CSSProperties = {
        border: "2px solid",
        borderRadius: "50%",
        cursor: "pointer",
        background: "transparent",
        color: "lightgray",
        padding: "10px",
    }
    
    const mediaOffBtn: React.CSSProperties = { 
        border: "2px solid",
        borderRadius: "50%",
        cursor: "pointer",
        background: "#D10000",
        color: "lightgray",
        padding: "10px",
    }

    const loadingStyle: React.CSSProperties = {
        position: "absolute", 
        top: 0,
        bottom: 0,
        left: 0,
        right: 0,
        margin: "auto",
    }


    return <>
    <div style={videoFrame}> 
        {loading && <CircularProgress style={loadingStyle} size="3rem"/>}
        <video ref={mediaRef} autoPlay playsInline style={videoBox}/>
        {!loading && <>
            <div style={mediaBar}>
                {audio && <MicIcon style={mediaOnBtn} onClick={() => toggleAudio()}/>}
                {!audio && <MicOffIcon style={mediaOffBtn} onClick={() => toggleAudio()}/>}

                {video && <VideocamIcon style={mediaOnBtn}  onClick={() => toggleVideo()}/>}
                {!video && <VideocamOffIcon style={mediaOffBtn}  onClick={() => toggleVideo()}/>}
            </div>
        </>
        }
    </div>

    </>
}