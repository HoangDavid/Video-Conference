import '../App.css';
import { useLocation } from "react-router-dom";
import { useState } from 'react';
import { Toaster, toast } from "react-hot-toast";
import CircularProgress from '@mui/material/CircularProgress';
import { useNavigate } from 'react-router-dom';
import PreviewVideo from '../components/preview';

import type { Room } from "../types/room";
import get_claims from '../services/security/getClaims';
import { signal_conn } from '../services/lib/signal';
import { pub_conn, waitUntilDualPcConnect } from '../services/lib/webrtc';
import { media } from '../services/lib/media';


export default function NewMeetingPage(){
    // Logic
    const state = useLocation().state as Room;
    const navigate = useNavigate();
    const [loading, setLoading] = useState(false);

    const onStartMeeting = async () => {
        setLoading(true)
        const claims = await get_claims();
        if (!claims) {toast.error("you are not authenticated");navigate("/");return;}

        if (!await signal_conn.waitUnillOpen()){toast.error("unable to connect to signaling server");setLoading(false);return;}

        // send start meeting request
        signal_conn.sendAction(claims, "start_room");
        const stream = await media.getAV();
        pub_conn.attachLocalStream(stream);

        if(!await waitUntilDualPcConnect(signal_conn)){toast.error("unable to connect to sfu server");setLoading(false); return;};
        navigate(`/rooms/${claims.roomID}/meeting`);
        setLoading(false);
    }

    // customization
    const box: React.CSSProperties = {
        display: "flex",
        justifyContent: "center",
    }

    return <>
    <h1>VideoCall Demo</h1>
    <h3> Check your video and audio before joining:</h3>
    <div style={box}>
        <PreviewVideo/>
        <div>
            <h3> Room Created!</h3>
            <div className='box' onClick={
                async() =>{
                    await navigator.clipboard.writeText(`roomID: ${state.roomID}\npin: ${state.pin}`)
                    toast.success("Copied")
            }}>
                <p>Room ID: {state.roomID}</p>
                <p>Pin: {state.pin}</p>
                <h5> (Click to copy)</h5>
            </div>

            <div className="card"> 
            {loading && <CircularProgress size="3rem"/>}
            {!loading &&
                <button onClick={async () => await onStartMeeting()}>
                    Start Meeting
                </button>
            }
            </div>
        </div>
    </div>
    <Toaster position="top-center" />
    </>
}