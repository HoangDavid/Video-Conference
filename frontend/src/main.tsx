// import { StrictMode } from 'react'
import { createBrowserRouter, RouterProvider } from "react-router-dom";
import { createRoot } from 'react-dom/client'
import NewMeetingPage from "./pages/NewMeetingPage.tsx";
import LobbyPage from "./pages/LobbyPage.tsx";
import MeetingPage from "./pages/LiveMeetingPage.tsx";
import PreviewPage from "./pages/PreviewPage.tsx"
import './index.css'
import App from './App.tsx'

export const router = createBrowserRouter([
  {path: "/", element: <App/>},
  {path: "/rooms/new", element: <NewMeetingPage/>},
  {path: "/rooms/:roomID/preview", element: <PreviewPage/>},
  {path: "/rooms/:roomID/lobby", element: <LobbyPage/>},
  {path: "/rooms/:roomID/meeting",  element: <MeetingPage/>}
])

createRoot(document.getElementById('root')!).render(
  <RouterProvider router={router}/>
)
