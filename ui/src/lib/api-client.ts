import type { GetSchemaResponse } from '@/types/telemetry'
import { API_BASE_URL } from '@/config/api'

interface ApiError extends Error {
  status?: number
}

class ApiError extends Error {
  constructor(message: string, public status?: number) {
    super(message)
    this.name = 'ApiError'
  }
}

const handleResponse = async <T>(response: Response): Promise<T> => {
  if (!response.ok) {
    const error = new ApiError('API request failed', response.status)
    throw error
  }
  return response.json()
}

export const apiClient = {
  get: async <T>(endpoint: string): Promise<T> => {
    const response = await fetch(`${API_BASE_URL}${endpoint}`)
    return handleResponse<T>(response)
  },

  post: async <T>(endpoint: string, data: unknown): Promise<T> => {
    const response = await fetch(`${API_BASE_URL}${endpoint}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    })
    return handleResponse<T>(response)
  },

  put: async <T>(endpoint: string, data: unknown): Promise<T> => {
    const response = await fetch(`${API_BASE_URL}${endpoint}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    })
    return handleResponse<T>(response)
  },

  delete: async <T>(endpoint: string): Promise<T> => {
    const response = await fetch(`${API_BASE_URL}${endpoint}`, {
      method: 'DELETE',
    })
    return handleResponse<T>(response)
  },
}

// Common types
interface ListResponse<T> {
  items: T[]
  total: number
  page: number
  pageSize: number
}

interface ListParams {
  page: number
  pageSize: number
  search?: string
  sortField?: string
  sortDirection?: 'asc' | 'desc'
  filters?: Record<string, string[]>
}

// Schema types
interface ListSchemasResponse extends ListResponse<GetSchemaResponse> {}
interface ListSchemasParams extends ListParams {
  type?: string
}

// API endpoints organized by domain
export const api = {
  schemas: {
    getByKey: (key: string) => apiClient.get<GetSchemaResponse>(`/api/v1/schemas/${key}`),
    list: () => apiClient.get<GetSchemaResponse[]>('/api/v1/schemas'),
    listWithParams: (params: ListSchemasParams) => {
      const searchParams = new URLSearchParams({
        page: params.page.toString(),
        pageSize: params.pageSize.toString(),
      })

      if (params.search) {
        searchParams.append('search', params.search)
      }
      if (params.type && params.type !== 'all') {
        searchParams.append('type', params.type)
      }
      if (params.sortField) {
        searchParams.append('sort_field', params.sortField)
      }
      if (params.sortDirection) {
        searchParams.append('sort_direction', params.sortDirection)
      }

      if (params.filters) {
        Object.entries(params.filters).forEach(([key, values]) => {
          if (values.length > 0) {
            values.forEach(value => {
              searchParams.append(`filter_${key}`, value)
            })
          }
        })
      }

      return apiClient.get<ListSchemasResponse>(`/api/v1/schemas?${searchParams.toString()}`)
    },
  },
  // Example of how to add new domains:
  // users: {
  //   getById: (id: string) => apiClient.get<User>(`/api/v1/users/${id}`),
  //   list: (params: ListParams) => apiClient.get<ListResponse<User>>('/api/v1/users'),
  //   create: (data: CreateUserRequest) => apiClient.post<User>('/api/v1/users', data),
  //   update: (id: string, data: UpdateUserRequest) => apiClient.put<User>(`/api/v1/users/${id}`, data),
  //   delete: (id: string) => apiClient.delete<void>(`/api/v1/users/${id}`),
  // },
  // teams: {
  //   getById: (id: string) => apiClient.get<Team>(`/api/v1/teams/${id}`),
  //   list: (params: ListParams) => apiClient.get<ListResponse<Team>>('/api/v1/teams'),
  //   create: (data: CreateTeamRequest) => apiClient.post<Team>('/api/v1/teams', data),
  // },
} 