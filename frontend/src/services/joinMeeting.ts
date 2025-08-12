

export default async function join(userName: string, roomID: string){

    const base_url = import.meta.env.VITE_SIGNALING_URL
    const res = await fetch(`https://${base_url}/rooms/${roomID}/auth?name=${encodeURIComponent(userName)}`, {method: "POST"})

    
    return res
}