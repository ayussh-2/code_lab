import { apiRequest } from "@/lib/api";

export interface PublicProfile {
    username: string;
    name: string;
    avatar_url: string;
    bio: string;
    created_at: string;
}

export interface OwnProfile extends PublicProfile {
    id: number;
    email: string;
    role: string;
}

export interface UserStats {
    solved_by_difficulty: {
        easy: number;
        medium: number;
        hard: number;
    };
    total_solved: number;
    acceptance_rate: number;
    total_submits: number;
    total_ac: number;
}

export interface ActivityDay {
    date: string;
    count: number;
}

export interface ProfileSubmission {
    id: number;
    problem_id: number;
    problem_slug: string;
    problem_title: string;
    language: string;
    verdict: string;
    status: string;
    runtime_ms: number;
    created_at: string;
}

export interface UpdateProfilePayload {
    name?: string;
    username?: string;
    avatar_url?: string;
    bio?: string;
}

export async function getOwnProfile() {
    return apiRequest<OwnProfile>("/profile/me");
}

export async function updateOwnProfile(body: UpdateProfilePayload) {
    return apiRequest<OwnProfile>("/profile/me", { method: "PATCH", body });
}

export async function getPublicProfile(username: string) {
    return apiRequest<PublicProfile>(`/users/${encodeURIComponent(username)}`);
}

export async function getUserStats(username: string) {
    return apiRequest<UserStats>(
        `/users/${encodeURIComponent(username)}/stats`,
    );
}

export async function getActivityHeatmap(username: string) {
    return apiRequest<ActivityDay[]>(
        `/users/${encodeURIComponent(username)}/activity`,
    );
}

export async function getProfileSubmissions(username: string, limit = 50) {
    return apiRequest<ProfileSubmission[]>(
        `/users/${encodeURIComponent(username)}/submissions?limit=${limit}`,
    );
}
