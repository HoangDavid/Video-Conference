import CircularProgress from '@mui/material/CircularProgress';
import { Toaster, toast } from "react-hot-toast";
import { getSignal } from '../services/lib/instance';


export default function Lobby(){

    const signal_conn = getSignal()
    
    
    // Logic
    return <>
    <h1> Lobby </h1>
    </>
}