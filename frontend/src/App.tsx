import './App.css';
import {useState } from 'react';
import React from "react";
import create_meeting from './services/createMeeting';
import join_meeting from './services/joinMeeting';


function App() {

  // Landing page logic

  const [name, setName] = useState("");
  const [roomID, setRoomID] = useState("")
  const [duration, setDuration] = useState("")
  const options = ["15m", "30m", "1h"] as const;
  type Option = typeof options[number];

  const [isJoin, setIsJoin] = useState(false);
  const [isCreateRoom, setIsCreateRoom] = useState(false)


  const createRoom = async (userName:string, duration: string) => {
  
  }

  const joinRoom = async (userName:string, roomID: string) => {

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
    fontSize: "15px",
    verticalAlign: "middle"
  }


  return (
    <>
      <h1>VideoCall Demo</h1>

      {/* Enter Name */}
      <input
      type="text"
      value={name}
      onChange={(e) => setName(e.target.value)}
      placeholder="Enter your name here"
      style={inputx}
      />

      {/* Enter RoomID */}
      {isJoin &&
        <input
        type="text"
        value={roomID}
        style={inputx}
        onChange={(e) => setRoomID(e.target.value)}
        placeholder="Enter RoomID"
        />
      }

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
              else console.log()
            }}>
              Create Room
            </button>
          }

          {!isCreateRoom &&
            <button style={btn} onClick={() => {
              if (!isJoin) setIsJoin(true)
              else console.log()
            }}>
              Join Room
            </button>
          }
          
      </div>
    </>
  )
}

export default App
