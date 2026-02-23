const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080";

interface HealthResponse {
  status: string;
  message: string;
  uptime: string;
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const lang = localStorage.getItem("i18nextLng") || "pt-BR";

  const response = await fetch(`${API_URL}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      "Accept-Language": lang,
      ...options?.headers,
    },
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: response.statusText }));
    throw new Error(error.message || response.statusText);
  }

  return response.json();
}

export const api = {
  health: () => request<HealthResponse>("/api/v1/health"),
};
