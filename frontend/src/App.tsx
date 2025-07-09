import './App.css'
import {useState } from 'react'
import OnePeerClient from './components/webrtc'


function App() {
  
  const [start, setStart] = useState<boolean>(false)

  return (
    <>
      <h1>VidCall demo</h1>
      <div className="card">
        {start && <OnePeerClient/>}

        {!start && (
        <>
          <button onClick={() => setStart(!start)}>
            Start Call
          </button>
        <></> 
          <button>
            Join Call
          </button>
        </>
        )}
        <p>
          Edit <code>src/App.tsx</code> and save to test HMR
        </p>
      </div>
      <p className="read-the-docs">
        Click on the Vite and React logos to learn more
      </p>
    </>
  )
}

export default App
