import { apiRequest } from "@/lib/api";
import {
    frontendQuestionsAsListItems,
    getFrontendQuestionBySlug,
} from "@/lib/frontend-questions";

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

function matchesSearch(item: ProblemListItem, query: string): boolean {
  if (!query) return true;
  const q = query.toLowerCase();
  return (
    item.title.toLowerCase().includes(q) ||
    item.slug.toLowerCase().includes(q) ||
    item.topics.some((topic) => topic.toLowerCase().includes(q))
  );
}

export async function getProblems(searchTerm = "") {
  const query = searchTerm.trim();
  const params = query ? `?search=${encodeURIComponent(query)}` : "";
  const result = await apiRequest<ProblemListItem[]>(`/problems${params}`);

  const hardcoded = frontendQuestionsAsListItems().filter((item) =>
    matchesSearch(item, query),
  );

  return {
    message: result.message,
    data: [...hardcoded, ...result.data],
  };
}

export async function getProblemBySlug(slug: string) {
  const hardcoded = getFrontendQuestionBySlug(slug);
  if (hardcoded) {
    return { message: "ok", data: hardcoded };
  }
  return apiRequest<ProblemDetail>(`/problems/${slug}`);
}
