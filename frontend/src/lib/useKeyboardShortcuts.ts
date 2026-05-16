import { useEffect } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'

/**
 * Global keyboard shortcuts for desktop app efficiency:
 * - Ctrl+N: New Patient (navigates to patients with action=new)
 * - Ctrl+B: New Invoice (navigates to billing with action=new)
 * - F2: Focus search input on current page
 * - Esc: Close open forms/dialogs (handled by individual components)
 */
export function useKeyboardShortcuts() {
  const navigate = useNavigate()
  const location = useLocation()

  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      // Don't trigger shortcuts when typing in inputs
      const target = e.target as HTMLElement
      const isInput = target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.tagName === 'SELECT'

      // Ctrl+N: New Patient
      if ((e.ctrlKey || e.metaKey) && e.key === 'n' && !e.shiftKey) {
        e.preventDefault()
        navigate('/patients?action=new')
        return
      }

      // Ctrl+B: New Invoice
      if ((e.ctrlKey || e.metaKey) && e.key === 'b' && !e.shiftKey) {
        e.preventDefault()
        navigate('/billing?action=new')
        return
      }

      // F2: Focus search input
      if (e.key === 'F2' && !isInput) {
        e.preventDefault()
        const searchInput = document.querySelector<HTMLInputElement>('[data-search-input]')
        if (searchInput) {
          searchInput.focus()
          searchInput.select()
        }
      }

      // Ctrl+P: Print (on invoice detail page)
      if ((e.ctrlKey || e.metaKey) && e.key === 'p' && location.pathname.startsWith('/billing/')) {
        // Let browser default print handle it
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [navigate, location.pathname])
}
