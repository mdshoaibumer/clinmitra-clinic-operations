/**
 * Centralized API error handling layer.
 * Normalizes all Wails backend errors into a consistent format
 * for the frontend stores and components.
 *
 * Time complexity: O(1) per call — simple try/catch wrapper
 * Space complexity: O(1) — no retained state
 */

export interface AppError {
  code: string
  message: string
}

/**
 * Parses a Wails error (which arrives as a string) into a structured AppError.
 * Backend errors follow the format "[CODE] Message" from utils.AppError.
 */
export function parseError(error: unknown): AppError {
  if (error === null || error === undefined) {
    return { code: 'UNKNOWN', message: 'An unexpected error occurred' }
  }

  const msg = typeof error === 'string' ? error : String(error)

  // Match backend AppError format: [CODE] Message
  const match = msg.match(/^\[([A-Z_]+)\]\s*(.+)$/)
  if (match) {
    return { code: match[1], message: match[2] }
  }

  return { code: 'UNKNOWN', message: msg || 'An unexpected error occurred' }
}

/**
 * Wraps an async Wails binding call with consistent error handling.
 * Returns [result, null] on success, [null, AppError] on failure.
 *
 * Usage:
 *   const [patient, err] = await apiCall(() => window.go.handler.PatientHandler.GetPatient(id))
 *   if (err) { showToast(err.message); return }
 */
export async function apiCall<T>(fn: () => Promise<T>): Promise<[T, null] | [null, AppError]> {
  try {
    const result = await fn()
    return [result, null]
  } catch (error) {
    return [null, parseError(error)]
  }
}

/**
 * Returns a user-friendly message for known error codes.
 */
export function getErrorMessage(error: AppError): string {
  switch (error.code) {
    case 'UNAUTHORIZED':
      return 'Please log in to continue'
    case 'FORBIDDEN':
      return 'You do not have permission to perform this action'
    case 'NOT_FOUND':
      return 'The requested resource was not found'
    case 'ACCOUNT_LOCKED':
      return 'Account is temporarily locked. Please try again later.'
    case 'VALIDATION_ERROR':
      return error.message
    case 'DUPLICATE':
      return error.message || 'This record already exists'
    case 'INTERNAL_ERROR':
      return 'Something went wrong. Please try again.'
    default:
      return error.message
  }
}
