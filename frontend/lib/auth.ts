import {
  API_BASE_URL,
  ApiError,
  apiRequest,
  apiRequestWithAuth,
  refreshAccessToken,
} from "@/lib/api";
import { clearSession, getAccessToken } from "@/lib/session";

export interface User {
  id: number;
  name: string;
  email: string;
  role: string;
}

export interface LoginResponse {
  access_token: string;
  user: User;
}

export interface RegisterResponse {
  id: number;
  name: string;
  email: string;
  role: string;
}

export interface LoginPayload {
  email: string;
  password: string;
}

export interface SignupPayload {
  name: string;
  email: string;
  password: string;
}

export function getGoogleLoginUrl(): string {
  return `${API_BASE_URL}/auth/google/login`;
}

export async function login(payload: LoginPayload) {
  return apiRequest<LoginResponse>("/auth/login", {
    method: "POST",
    body: payload,
  });
}

export async function signup(payload: SignupPayload) {
  return apiRequest<RegisterResponse>("/auth/register", {
    method: "POST",
    body: payload,
  });
}

export async function verifyEmail(payload: { email: string; code: string }) {
  return apiRequest<null>("/auth/verify-email", {
    method: "POST",
    body: payload,
  });
}

export async function resendVerificationEmail(email: string) {
  return apiRequest<null>("/auth/verify-email/resend", {
    method: "POST",
    body: { email },
  });
}

export async function requestPasswordReset(email: string) {
  return apiRequest<null>("/auth/password-reset/request", {
    method: "POST",
    body: { email },
  });
}

export async function confirmPasswordReset(payload: {
  email: string;
  code: string;
  new_password: string;
}) {
  return apiRequest<null>("/auth/password-reset/confirm", {
    method: "POST",
    body: payload,
  });
}

export async function logout() {
  const token = getAccessToken();
  try {
    if (token) {
      await apiRequestWithAuth<null>("/auth/logout", token, { method: "POST" });
    }
  } catch (error) {
    if (!(error instanceof ApiError) || error.status !== 401) {
      throw error;
    }
  } finally {
    clearSession();
  }
}

export async function ensureSession(): Promise<boolean> {
  if (getAccessToken()) return true;
  const token = await refreshAccessToken();
  return token !== null;
}
