
export default async function authenticate(userName: string, roomID: string, pin: string): Promise<boolean>{

    const res = await fetch(
        `/api/rooms/${roomID}/auth?name=${encodeURIComponent(userName)}`, 
        {
            method: "POST",
            credentials: "include",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ pin: pin}),
        }
        
    );

    if (!res.ok) {
        const msg = await res.text().catch(() => "");
        console.log(`HTTP ${res.status} ${msg}`)
        return false
    }
    
    return true
}