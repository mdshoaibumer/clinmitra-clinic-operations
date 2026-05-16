import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import './index.css'

console.log('VITE_USE_MOCKS:', import.meta.env.VITE_USE_MOCKS)
if (import.meta.env.VITE_USE_MOCKS === 'true') {
  console.log('Loading Wails Mocks...')
  await import('./lib/wails-mock')
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
)
