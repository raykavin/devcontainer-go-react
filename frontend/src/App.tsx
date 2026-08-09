import { useState } from 'react'
import './App.css'

function App() {
  const [count, setCount] = useState(0)

  return (
    <section id="center">
      <h1>Get started</h1>
      <p>
        Edit <code>src/App.tsx</code> and save to test <code>HMR</code>
      </p>
      <button type="button" className="counter" onClick={() => setCount((c) => c + 1)}>
        Count is {count}
      </button>
      <p className="hint">
        API calls to <code>/api</code> are proxied to the Go backend on{' '}
        <code>:8080</code> — see <code>vite.config.ts</code>.
      </p>
    </section>
  )
}

export default App
