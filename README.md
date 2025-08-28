# Video Conference

### Currently a WIP

A work-in-progress **video conferencing system** built around an **SFU (Selective Forwarding Unit)** architecture.  
It uses **WebRTC** for real-time audio/video, **WebSockets** and **gRPC** for signaling, **MongoDB** for room and user state, and **cookie-based authentication** for secure session management.  
Designed to support multi-party meetings with efficient media routing and a scalable backend.

---

## Features
- Multi-party video conferencing using WebRTC and SFU
- Real-time **signaling** over WebSockets and gRPC
- **MongoDB** for persisting room and user state
- **Cookie-based authentication**
- Frontend built with **React**
- Backend written in **Go**

---

## Prerequisites
Make sure you have the following installed on your system:
- [Go](https://go.dev/) (≥ 1.20)
- [Node.js](https://nodejs.org/) and [Yarn](https://yarnpkg.com/)
- [MongoDB](https://www.mongodb.com/try/download/community) running locally

---

## How to Run

### Backend
1. Update your `.env` in `/backend`:
   - Run a local **MongoDB** instance.
   - Add its **connection URL** and **database name** to `.env`.
   - Configure the ports you want for the **signaling** and **SFU** servers.

2. Start the backend services:
   ```bash
   cd backend
   go run cmd/signaling/main.go
   go run cmd/sfu/main.go
   ```

### Frontend
Start frontend:
```bash
cd frontend
yarn install
yarn dev
```

### Secured Backend (optional):
```bash
mkcert -install
mkcert -key-file dev.key -cert-file dev.crt
```
Then update your key and cert url in your .env file in /backend for TLS configuration

### Multi devices testing (optional):
To test on other devices, expose your local frontend with a free tunneling service:
```bash
cloudflared tunnel --url [your local frontend url]
```

