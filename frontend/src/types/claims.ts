export type Role = "host" | "guest";

export interface Claims {
    name: string
    ID: string
    roomID: string
    role: Role
}