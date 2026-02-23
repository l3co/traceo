import { auth } from "@/shared/lib/firebase";

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080";

export interface HealthResponse {
  status: string;
  message: string;
  uptime: string;
}

export interface UserResponse {
  id: string;
  name: string;
  email: string;
  phone?: string;
  cell_phone?: string;
  avatar_url?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateUserInput {
  name: string;
  email: string;
  password: string;
  phone?: string;
  cell_phone?: string;
  accepted_terms: boolean;
}

export interface UpdateUserInput {
  name: string;
  phone?: string;
  cell_phone?: string;
}

async function getAuthHeader(): Promise<Record<string, string>> {
  const user = auth.currentUser;
  if (!user) return {};
  const token = await user.getIdToken();
  return { Authorization: `Bearer ${token}` };
}

async function request<T>(
  path: string,
  options?: RequestInit & { skipAuth?: boolean }
): Promise<T> {
  const lang = localStorage.getItem("i18nextLng") || "pt-BR";
  const authHeaders = options?.skipAuth ? {} : await getAuthHeader();

  const response = await fetch(`${API_URL}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      "Accept-Language": lang,
      ...authHeaders,
      ...options?.headers,
    },
  });

  if (response.status === 204) return undefined as T;

  if (!response.ok) {
    const error = await response
      .json()
      .catch(() => ({ message: response.statusText }));
    throw new Error(error.message || response.statusText);
  }

  return response.json();
}

// --- Missing Person Types ---

export interface MissingResponse {
  id: string;
  user_id: string;
  name: string;
  nickname?: string;
  birth_date?: string;
  date_of_disappearance?: string;
  height?: string;
  clothes?: string;
  gender: string;
  eyes: string;
  hair: string;
  skin: string;
  photo_url?: string;
  lat: number;
  lng: number;
  status: string;
  event_report?: string;
  tattoo_description?: string;
  scar_description?: string;
  was_child: boolean;
  slug: string;
  has_tattoo: boolean;
  has_scar: boolean;
  age: number;
  created_at: string;
  updated_at: string;
}

export interface MissingListResponse {
  items: MissingResponse[];
  next_cursor?: string;
}

export interface CreateMissingInput {
  name: string;
  nickname?: string;
  birth_date?: string;
  date_of_disappearance?: string;
  height?: string;
  clothes?: string;
  gender: string;
  eyes: string;
  hair: string;
  skin: string;
  photo_url?: string;
  lat: number;
  lng: number;
  event_report?: string;
  tattoo_description?: string;
  scar_description?: string;
}

export interface UpdateMissingInput extends CreateMissingInput {
  status?: string;
}

export interface GenderStatDTO {
  gender: string;
  count: number;
}

export interface YearStatDTO {
  year: number;
  count: number;
}

export interface StatsResponse {
  total: number;
  by_gender: GenderStatDTO[];
  child_count: number;
  by_year: YearStatDTO[];
}

export interface LocationPointDTO {
  id: string;
  name: string;
  lat: number;
  lng: number;
  status: string;
}

export interface LocationsResponse {
  locations: LocationPointDTO[];
}

export const api = {
  health: () =>
    request<HealthResponse>("/api/v1/health", { skipAuth: true }),

  createUser: (data: CreateUserInput) =>
    request<UserResponse>("/api/v1/users", {
      method: "POST",
      body: JSON.stringify(data),
      skipAuth: true,
    }),

  getUser: (id: string) => request<UserResponse>(`/api/v1/users/${id}`),

  updateUser: (id: string, data: UpdateUserInput) =>
    request<UserResponse>(`/api/v1/users/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  deleteUser: (id: string) =>
    request<void>(`/api/v1/users/${id}`, { method: "DELETE" }),

  changePassword: (id: string, newPassword: string) =>
    request<void>(`/api/v1/users/${id}/password`, {
      method: "PATCH",
      body: JSON.stringify({ new_password: newPassword }),
    }),

  forgotPassword: (email: string) =>
    request<void>("/api/v1/auth/forgot-password", {
      method: "POST",
      body: JSON.stringify({ email }),
      skipAuth: true,
    }),

  // --- Missing Persons ---

  listMissing: (size = 20, after?: string) => {
    const params = new URLSearchParams({ size: String(size) });
    if (after) params.set("after", after);
    return request<MissingListResponse>(
      `/api/v1/missing?${params.toString()}`,
      { skipAuth: true }
    );
  },

  getMissing: (id: string) =>
    request<MissingResponse>(`/api/v1/missing/${id}`, { skipAuth: true }),

  createMissing: (data: CreateMissingInput) =>
    request<MissingResponse>("/api/v1/missing", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  updateMissing: (id: string, data: UpdateMissingInput) =>
    request<MissingResponse>(`/api/v1/missing/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  deleteMissing: (id: string) =>
    request<void>(`/api/v1/missing/${id}`, { method: "DELETE" }),

  getUserMissing: (userId: string) =>
    request<MissingResponse[]>(`/api/v1/users/${userId}/missing`),

  searchMissing: (q: string, limit = 20) =>
    request<MissingResponse[]>(
      `/api/v1/missing/search?q=${encodeURIComponent(q)}&limit=${limit}`,
      { skipAuth: true }
    ),

  getMissingStats: () =>
    request<StatsResponse>("/api/v1/missing/stats", { skipAuth: true }),

  getMissingLocations: (limit = 100) =>
    request<LocationsResponse>(`/api/v1/missing/locations?limit=${limit}`, {
      skipAuth: true,
    }),
};
