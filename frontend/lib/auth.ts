import { API_BASE_URL, apiRequest } from "@/lib/api";

export interface User {
  id: number;
  name: string;
  email: string;
  role: string;
}

export interface LoginResponse {
  user: User;
}

export interface MeResponse {
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

export async function me() {
  return apiRequest<MeResponse>("/auth/me");
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
  await apiRequest<null>("/auth/logout", { method: "POST" });
}
