// import { StrictMode } from 'react'
import { createBrowserRouter, RouterProvider } from "react-router-dom";
import { createRoot } from 'react-dom/client'
import NewMeeting from "./pages/NewMeetingPage.tsx";
import Lobby from "./pages/LobbyPage.tsx";
import './index.css'
import App from './App.tsx'

const router = createBrowserRouter([
  {path: "/", element: <App/>},
  {path: "/rooms/new", element: <NewMeeting/>},
  {path: "/rooms/:roomID/lobby", element: <Lobby/>}
])

createRoot(document.getElementById('root')!).render(
  <RouterProvider router={router}/>
)
