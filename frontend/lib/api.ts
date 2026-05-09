import { clearSession, getAccessToken, saveSession } from "@/lib/session";

export interface ApiEnvelope<T> {
  error: boolean;
  message: string;
  data?: T;
}

export class ApiError extends Error {
  status: number;
  details: unknown;

  constructor(message: string, status: number, details?: unknown) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.details = details ?? null;
  }
}

export const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080/api";

interface ApiRequestOptions extends Omit<RequestInit, "body"> {
  body?: unknown;
}

export async function apiRequest<T>(
  path: string,
  options: ApiRequestOptions = {},
): Promise<{ message: string; data: T }> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(options.headers ?? {}),
    },
    body: options.body === undefined ? undefined : JSON.stringify(options.body),
    credentials: "include",
  });

  let payload: ApiEnvelope<T> | null = null;
  try {
    payload = (await response.json()) as ApiEnvelope<T>;
  } catch {
    throw new ApiError("Invalid API response", response.status);
  }

  if (!response.ok || payload.error) {
    throw new ApiError(
      payload.message || "Request failed",
      response.status,
      payload.data,
    );
  }

  return {
    message: payload.message,
    data: payload.data as T,
  };
}

interface RefreshPayload {
  access_token: string;
  user: { id: number; name: string; email: string; role: string };
}

let refreshInFlight: Promise<string | null> | null = null;

export function refreshAccessToken(): Promise<string | null> {
  if (refreshInFlight) return refreshInFlight;

  refreshInFlight = (async () => {
    try {
      const result = await apiRequest<RefreshPayload>("/auth/refresh", {
        method: "POST",
      });
      saveSession(result.data.access_token, result.data.user);
      return result.data.access_token;
    } catch {
      clearSession();
      return null;
    } finally {
      refreshInFlight = null;
    }
  })();

  return refreshInFlight;
}

async function fetchWithToken<T>(
  path: string,
  token: string,
  options: ApiRequestOptions,
): Promise<{ message: string; data: T }> {
  return apiRequest<T>(path, {
    ...options,
    headers: {
      ...(options.headers as Record<string, string> | undefined),
      Authorization: `Bearer ${token}`,
    },
  });
}

export async function apiRequestWithAuth<T>(
  path: string,
  token: string | null,
  options: ApiRequestOptions = {},
): Promise<{ message: string; data: T }> {
  let activeToken = token ?? getAccessToken();

  if (!activeToken) {
    activeToken = await refreshAccessToken();
    if (!activeToken) {
      throw new ApiError("Not authenticated", 401);
    }
  }

  try {
    return await fetchWithToken<T>(path, activeToken, options);
  } catch (error) {
    if (!(error instanceof ApiError) || error.status !== 401) {
      throw error;
    }

    const refreshed = await refreshAccessToken();
    if (!refreshed) throw error;

    return fetchWithToken<T>(path, refreshed, options);
  }
}
