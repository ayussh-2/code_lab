import { apiRequest } from "@/lib/api";

export interface ProblemListItem {
  id: number;
  title: string;
  slug: string;
  difficulty: "easy" | "medium" | "hard";
  topics: string[];
}

export interface ProblemDetail {
  id: number;
  title: string;
  slug: string;
  difficulty: "easy" | "medium" | "hard";
  topics: string[];
  hints: string[];
  details: string;
  examples: Array<{
    input: string;
    output: string;
    explanation: string;
  }>;
  constraints: string[];
  sample_test_cases: Array<{
    id: number;
    input: string;
    expected: string;
  }>;
}

export async function getProblems(searchTerm = "") {
  const query = searchTerm.trim();
  const params = query ? `?search=${encodeURIComponent(query)}` : "";
  return apiRequest<ProblemListItem[]>(`/problems${params}`);
}

export async function getProblemBySlug(slug: string) {
  return apiRequest<ProblemDetail>(`/problems/${slug}`);
}
