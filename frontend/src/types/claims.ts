export type Role = "host" | "guest";

export interface Claims {
    name: string
    peerID: string
    roomID: string
    role: Role
}