
export default async function create_meeting(userName: string, duration: string){
    
    const base_url = import.meta.env.VITE_SIGNALING_URL

    try{
        await fetch(`https://${base_url}/rooms/new/${duration}?name=${encodeURIComponent(userName)}`,
            {
                method: "GET",
                credentials: "include"
            }
        );
    }catch (e) {
        console.error(e)
    }

    // TODO: get a connetion to websocket here
    

    return 
}