import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import './index.css'

if (import.meta.env.VITE_USE_MOCKS === 'true') {
  await import('./lib/wails-mock')
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
)
