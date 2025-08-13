import type { Room } from "../types/room";

export default async function create_meeting(userName: string, duration: string): Promise<Room | null>{

    const res =  await fetch(
        `/api/rooms/new/${duration}?name=${encodeURIComponent(userName)}`,
        {
            method: "GET",
            credentials: "include"
        }
    );

    if (!res.ok){
        const msg = await res.text().catch(() => "");
        console.log(`HTTP ${res.status} ${msg}`)
        return null
    }

    const payload = (await res.json()) as Room;

    return payload
}