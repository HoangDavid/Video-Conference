import '../App.css';
import PreviewVideo from "../components/preview";
import { useNavigate } from 'react-router-dom';
import { Toaster, toast } from "react-hot-toast";
import { useState } from 'react';
import CircularProgress from '@mui/material/CircularProgress';

import get_claims from '../services/security/getClaims';
import { signal_conn } from '../services/lib/signal';
import {waitUntilDualPcConnect, pcConnect } from '../services/lib/webrtc';
import { media } from '../services/lib/media';
import type { PeerEvent } from '../types/signal';

export default function PreviewPage(){

    // Logic
    const navigate = useNavigate();
    const [loading, setLoading] = useState<boolean>(false)

    const onJoinMeeting = async() => {
        setLoading(true)
        const claims = await get_claims();
        if (!claims){ toast.error("you are not authenticated!"); navigate("/"); return;}

        signal_conn.connect()
        if (! await signal_conn.waitUnillOpen()) {toast.error("unable to connect to signal server"); setLoading(false); return;}
        
        // add callbacks for room live or not
        const del = signal_conn.onEvent(async (e: PeerEvent) => {

            if (!await waitUntilDualPcConnect()) { del(); return;}
            console.log(e.type)
            switch (e.type) {
            case "room_active":
                del();
                navigate(`../rooms/${claims.roomID}/meeting`);
                break;
            case "room_inactive":
                console.log("room inactive")
                del();
                navigate(`../rooms/${claims.roomID}/lobby`);
                break;
            }
        })

        // send join meeting action
        signal_conn.sendAction("join");

        // wait until pc connects
        const stream = await media.getAV();
        if (!await pcConnect(signal_conn, stream)) {toast.error("unable to connect to sfu server"); setLoading(false); return;}
    }

    // Customization
    const btn: React.CSSProperties = {
        marginTop: "10px"
    }

    return <>
        <h1>Preview</h1>
        <h3>Check your video and audio before joining</h3>
        <div>
            <PreviewVideo/>
            {loading && <CircularProgress size="3rem"/>}
            
            {!loading && <button style={btn} onClick={async() => await onJoinMeeting()}> Join Meeting</button>}
        </div>
        <Toaster position="top-center" />
    </>
}