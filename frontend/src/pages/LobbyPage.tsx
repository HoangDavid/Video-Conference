import '../App.css';
import CircularProgress from '@mui/material/CircularProgress';
import { useEffect, useState } from 'react';
import { Toaster, toast } from "react-hot-toast";
import { useNavigate } from 'react-router-dom';

import get_claims from '../services/security/getClaims';
import { signal_conn } from '../services/lib/signal';
import type { PeerEvent } from '../types/signal';

export default function LobbyPage(){

    // Logic:
    const navigate = useNavigate();
    const [loading, setLoading] = useState<boolean>(true)

    useEffect(() => {
        (async() => {
            if (await signal_conn.waitUnillOpen()) {
                const claims = await get_claims()
                if (!claims){
                    toast.error("You are not authenticated");
                    navigate("/");
                    return;
                }

                const del = signal_conn.onEvent((e: PeerEvent) => {
                    if (e.type == "room_active") {
                        setLoading(false);
                        del();
                        signal_conn.sendAction(claims, "join");
                        navigate(`rooms/${e.roomID}/meeting`);
                    }
                })
            }else {
                toast.error("unable to connect to signaling server");
            }


        })()
    })
    

    return <>
    <h1> Lobby </h1>
    <div>
        <div style={{marginTop: "10px"}}>
            {loading && <>
                <CircularProgress size="3rem"/>
                <p> Wait for your host to start the meeting</p>
            </>}

            {!loading && <p> Joining in 3..2..1</p>}  
        </div>
    </div>
    <Toaster position="top-center" />
    </>
}