import type {Claims}  from "../types/claims";

export default async function join_meeting(): Promise<boolean>{
    const claims = await fetch("/api/me",  {
            method: "GET",
            credentials: "include"
        });

    if (!claims.ok) {
        const msg = await claims.text().catch(() => "");
        console.log(`HTTP ${claims.status} ${msg}`)
        return false
    }

    const data = await claims.json() as Claims
    // TODO: connect to the websocket
    // TODO: establish peer connection
    if (data.role == 'host') {
        //  TODO: send a start meeting request
    }else if (data.role == 'guest') {
        //  TODO: send a join meeting request: if room not active -> lobby else go straight to room
    }


    return true

}