import './App.css';
import {useState } from 'react';
import React from "react";
import CircularProgress from '@mui/material/CircularProgress';
import { useNavigate } from 'react-router-dom';
import { Toaster, toast } from "react-hot-toast";

import create_meeting from './services/createMeeting';
import join_meeting from './services/joinMeeting';
import authenticate from './services/authenticate';



function App() {

  // Logic

  // for text field
  const [name, setName] = useState("");
  const [roomID, setRoomID] = useState("")
  const [pin, setPin] = useState("")
  const [duration, setDuration] = useState("")

  // for drop down
  const options = ["15m", "30m", "1h"] as const;
  type Option = typeof options[number];

  const [isJoin, setIsJoin] = useState(false);
  const [isCreateRoom, setIsCreateRoom] = useState(false)

  const [loading, setLoading] = useState(false)


  const navigate = useNavigate();

  const onCreateRoom = async (userName:string, duration: string) => {
    setLoading(true);
    const room = await create_meeting(userName, duration);
    navigate("/rooms/new", {state: {roomID: room?.roomID, pin: room?.pin}});
    setLoading(false);
  }

  const onJoinRoom = async (userName:string, roomID: string, pin: string) => {
    setLoading(true);
    if (await authenticate(userName, roomID, pin)) {
      if (!await join_meeting()) {
        toast.error("Unable to join meeting");
        setLoading(false);
      }

    }else {
      toast.error("Invalid room ID and/or pin");
      setLoading(false);
    }
  }


  // Customization
  const btn = {
    margin: "10px"
  };

  const inputx: React.CSSProperties = {
    padding: "10px 25px",
    borderRadius: "15px",
    fontWeight: "500",
    fontSize: "18px",
    display:"block",
    marginBottom: "20px",
    marginLeft: "auto",
    marginRight: "auto",
    textAlign: "center"
  };
  

  const dropdownx: React.CSSProperties = {
    borderRadius: "15px",
    padding: "12px",
    fontWeight: "500",
    fontSize: "18px",
    verticalAlign: "middle"
  }


  return (
    <>
      <h1>VideoCall Demo</h1>
      
      {loading && <CircularProgress size="3rem" />}
      
      {!loading &&
      <>
        {/* Enter Name */}
        <input
        type="text"
        value={name}
        onChange={(e) => setName(e.target.value)}
        placeholder="Enter your name here"
        style={inputx}
        id="name"
        autoComplete='off'
        />

        {/* Enter roomID */}
        {isJoin &&
          <input
          type="text"
          value={roomID}
          style={inputx}
          onChange={(e) => setRoomID(e.target.value)}
          placeholder="Enter RoomID"
          id="roomID"
          autoComplete='off'
          />
        }

        {/* Enter room pin */}
        {isJoin &&
        <input
        type="password"
        value={pin}
        style={inputx}
        onChange={(e) => setPin(e.target.value)}
        placeholder="Enter your room pin"
        id="pin"
        autoComplete='off'
        />}

        {/* Enter Duration */}
        {isCreateRoom &&
          <select
            value={duration}
            style={dropdownx}
            onChange={(e) => setDuration(e.target.value as Option | "")}>

            <option value="" disabled>Pick duration</option>
            {options.map((opt) => (
              <option key={opt} value={opt}>{opt}</option>
            ))}

          </select>
        }

        <div className="card">

            {(isJoin || isCreateRoom) &&
              <button style={btn} onClick={() => {
                setIsCreateRoom(false)
                setIsJoin(false)
              }}>
                Back
              </button>
            }

            {!isJoin && 
              <button style={btn} onClick={() => {
                if (!isCreateRoom) setIsCreateRoom(true)
                else onCreateRoom(name, duration)
              }}>
                Create Room
              </button>
            }

            {!isCreateRoom &&
              <button style={btn} onClick={() => {
                if (!isJoin) setIsJoin(true)
                else onJoinRoom(name, roomID, pin)
              }}>
                Join Room
              </button>
            }
        </div>
        <Toaster position="top-center" />
      </>
      }
    </>
  )
}

export default App
