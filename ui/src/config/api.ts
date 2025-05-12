const isDevelopment = import.meta.env.DEV

export const API_BASE_URL = isDevelopment ? 'http://localhost:8080' : ''

export const API_ENDPOINTS = {
  schemas: `${API_BASE_URL}/api/v1/schemas`,
} as const 