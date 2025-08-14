// import { StrictMode } from 'react'
import { createBrowserRouter, RouterProvider } from "react-router-dom";
import { createRoot } from 'react-dom/client'
import NewMeeting from "./pages/NewMeetingPage.tsx";
import Lobby from "./pages/LobbyPage.tsx";
import Meeting from "./pages/MeetingPaage.tsx";
import './index.css'
import App from './App.tsx'

export const router = createBrowserRouter([
  {path: "/", element: <App/>},
  {path: "/rooms/new", element: <NewMeeting/>},
  {path: "/rooms/:roomID/lobby", element: <Lobby/>},
  {path: "/rooms/:roomID/meeting",  element: <Meeting/>}
])

createRoot(document.getElementById('root')!).render(
  <RouterProvider router={router}/>
)
