import '../App.css';
import { useLocation } from "react-router-dom";
import { useState } from 'react';
import { Toaster, toast } from "react-hot-toast";
import CircularProgress from '@mui/material/CircularProgress';
import type { Room } from "../types/room";

import join_meeting from '../services/joinMeeting';

export default function NewMeeting(){
    // Logic
    const state = useLocation().state as Room
    const [loading, setLoading] = useState(false)

    const onStartMeeting = async () => {
        setLoading(true)
        if (!await join_meeting()) {
            toast.error("unable to join meetinyg")
            setLoading(false)
        }
    }

    return <>
    <h1>VideoCall Demo</h1>
    
    {loading && <CircularProgress size="3rem" />}

    {!loading && 
        <>
        <h3>Room Created!</h3>
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
            <button onClick={async () => await onStartMeeting()}>
                Start Meeting
            </button>
        </div>
        <Toaster position="top-center" />
        </>}
    </>
}