import { apiRequest } from "@/lib/api";

export type SubmissionStatus = "queued" | "running" | "done";
export type SubmissionKind = "submit" | "run";

export const TERMINAL_STATUS: SubmissionStatus = "done";

export const VERDICTS = [
    "PENDING",
    "AC",
    "WA",
    "TLE",
    "MLE",
    "CE",
    "RE",
    "IE",
] as const;
export type Verdict = (typeof VERDICTS)[number];

export interface SubmissionTestResult {
    id: number;
    submission_id: number;
    test_case_id: number;
    verdict: Verdict;
    runtime_ms: number;
    memory_kb: number;
    actual_output_preview?: string;
    created_at: string;
}

export interface Submission {
    id: number;
    user_id: number;
    problem_id: number;
    language: string;
    kind: SubmissionKind;
    source_code: string;
    status: SubmissionStatus;
    verdict: Verdict;
    runtime_ms: number;
    memory_kb: number;
    score: number;
    failed_test_case_id?: number | null;
    failed_input_preview?: string;
    failed_expected_preview?: string;
    failed_actual_preview?: string;
    compiler_output?: string;
    stderr_preview?: string;
    error?: string;
    results?: SubmissionTestResult[];
    created_at: string;
    updated_at: string;
    judged_at?: string | null;
}

export interface CreateSubmissionPayload {
    problem_slug: string;
    language: string;
    source_code: string;
    kind: SubmissionKind;
}

export interface CreateSubmissionResponse {
    submission_id: number;
}

export async function createSubmission(payload: CreateSubmissionPayload) {
    return apiRequest<CreateSubmissionResponse>(`/submissions`, {
        method: "POST",
        body: payload,
    });
}

export async function getSubmission(id: number) {
    return apiRequest<Submission>(`/submissions/${id}`);
}

export async function listSubmissions(params?: {
    problem_slug?: string;
    kind?: SubmissionKind;
    limit?: number;
}) {
    const q = new URLSearchParams();
    if (params?.problem_slug) q.set("problem_slug", params.problem_slug);
    if (params?.kind) q.set("kind", params.kind);
    if (params?.limit) q.set("limit", String(params.limit));
    const qs = q.toString();
    return apiRequest<Submission[]>(`/submissions${qs ? `?${qs}` : ""}`);
}

export function isTerminal(status: SubmissionStatus): boolean {
    return status === TERMINAL_STATUS;
}

export const VERDICT_LABEL: Record<Verdict, string> = {
    PENDING: "Pending",
    AC: "Accepted",
    WA: "Wrong Answer",
    TLE: "Time Limit Exceeded",
    MLE: "Memory Limit Exceeded",
    CE: "Compilation Error",
    RE: "Runtime Error",
    IE: "Internal Error",
};

export const VERDICT_BADGE_CLASS: Record<Verdict, string> = {
    PENDING: "bg-white/6 text-zinc-300",
    AC: "bg-[#1cbf73]/15 text-[#1cbf73]",
    WA: "bg-[#ef4743]/15 text-[#ef4743]",
    TLE: "bg-[#ef4743]/15 text-[#ef4743]",
    MLE: "bg-[#ef4743]/15 text-[#ef4743]",
    CE: "bg-[#e8a24a]/15 text-[#e8a24a]",
    RE: "bg-[#ef4743]/15 text-[#ef4743]",
    IE: "bg-zinc-500/15 text-zinc-300",
};
