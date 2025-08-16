import type {Claims}  from "../../types/claims";

export default async function get_claims(): Promise<Claims | null>{
    const res = await fetch("/api/me",  {
            method: "GET",
            credentials: "include"
        });

    if (!res.ok) {
        const msg = await res.text().catch(() => "");
        console.log(`HTTP ${res.status} ${msg}`)
        return null
    }

    const claims = await res.json() as Claims
    return claims
}