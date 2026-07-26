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
// update
export const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080/api";

interface ApiRequestOptions extends Omit<RequestInit, "body"> {
  body?: unknown;
}

async function rawFetch(path: string, options: ApiRequestOptions) {
  return fetch(`${API_BASE_URL}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(options.headers ?? {}),
    },
    body: options.body === undefined ? undefined : JSON.stringify(options.body),
    credentials: "include",
  });
}

async function parseEnvelope<T>(
  response: Response,
): Promise<{ message: string; data: T }> {
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

let refreshInFlight: Promise<boolean> | null = null;

async function tryRefresh(): Promise<boolean> {
  if (refreshInFlight) return refreshInFlight;
  refreshInFlight = (async () => {
    try {
      const response = await rawFetch("/auth/refresh", { method: "POST" });
      return response.ok;
    } catch {
      return false;
    } finally {
      refreshInFlight = null;
    }
  })();
  return refreshInFlight;
}

export async function apiRequest<T>(
  path: string,
  options: ApiRequestOptions = {},
): Promise<{ message: string; data: T }> {
  let response = await rawFetch(path, options);

  const isAuthEndpoint =
    path.startsWith("/auth/refresh") ||
    path.startsWith("/auth/login") ||
    path.startsWith("/auth/register");

  if (response.status === 401 && !isAuthEndpoint) {
    const refreshed = await tryRefresh();
    if (refreshed) {
      response = await rawFetch(path, options);
    }
  }

  return parseEnvelope<T>(response);
}
