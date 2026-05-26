import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'https://api.webhook.inst.lk';

export const api = axios.create({
  baseURL: API_BASE_URL + '/api/v1',
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor to attach access token
api.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = typeof window !== 'undefined' ? localStorage.getItem('access_token') : null;
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Response interceptor for token refresh
api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean };

    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      try {
        const { data } = await api.post('/auth/refresh');
        localStorage.setItem('access_token', data.access_token);
        originalRequest.headers.Authorization = `Bearer ${data.access_token}`;
        return api(originalRequest);
      } catch (refreshError) {
        localStorage.removeItem('access_token');
        if (typeof window !== 'undefined') {
          window.location.href = '/login';
        }
        return Promise.reject(refreshError);
      }
    }

    return Promise.reject(error);
  }
);

// Auth API
export const authAPI = {
  register: (data: { email: string; password: string; display_name: string }) =>
    api.post('/auth/register', data),
  login: (data: { email: string; password: string }) =>
    api.post('/auth/login', data),
  logout: () => api.post('/auth/logout'),
  refresh: () => api.post('/auth/refresh'),
  me: () => api.get('/me'),
};

// Endpoints API
export const endpointsAPI = {
  create: (data: { name?: string; description?: string; custom_token?: string }) =>
    api.post('/endpoints', data),
  list: (params?: { limit?: number; offset?: number }) =>
    api.get('/endpoints', { params }),
  get: (id: string) => api.get(`/endpoints/${id}`),
  update: (id: string, data: { name?: string; description?: string }) =>
    api.put(`/endpoints/${id}`, data),
  delete: (id: string) => api.delete(`/endpoints/${id}`),
  setCustomResponse: (id: string, data: {
    status_code: number;
    headers?: Record<string, string>;
    body?: string;
    delay?: number;
  }) => api.put(`/endpoints/${id}/response`, data),
};

// Requests API
export const requestsAPI = {
  list: (endpointId: string, params?: { limit?: number; offset?: number; search?: string }) =>
    api.get(`/endpoints/${endpointId}/requests`, { params }),
  get: (endpointId: string, requestId: string) =>
    api.get(`/endpoints/${endpointId}/requests/${requestId}`),
  delete: (endpointId: string, requestId: string) =>
    api.delete(`/endpoints/${endpointId}/requests/${requestId}`),
  replay: (endpointId: string, requestId: string, data: {
    method?: string;
    url: string;
    headers?: Record<string, string>;
    body?: string;
  }) => api.post(`/endpoints/${endpointId}/requests/${requestId}/replay`, data),
};

// API Keys API
export const apiKeysAPI = {
  create: (data: { name: string; scopes: string[]; expires_in?: number }) =>
    api.post('/api-keys', data),
  list: () => api.get('/api-keys'),
  revoke: (id: string) => api.delete(`/api-keys/${id}`),
};
