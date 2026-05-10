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
  topic_ids?: number[];
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

export interface TopicRow {
  id: number;
  name: string;
}

export async function listTopics() {
  return apiRequest<TopicRow[]>(`/problems/topics`);
}

export interface CreateProblemPayload {
  title: string;
  difficulty: "easy" | "medium" | "hard";
  topics: number[];
  hints: string[];
  details: string;
  examples: Array<{
    input: string;
    output: string;
    explanation: string;
  }>;
  constraints: string[];
  sample_test_cases: Array<{
    input: string;
    expected: string;
  }>;
}

export interface CreatedProblem {
  id: number;
  title: string;
  slug: string;
  difficulty: string;
}

export async function createProblem(body: CreateProblemPayload) {
  return apiRequest<CreatedProblem>(`/problems`, {
    method: "POST",
    body,
  });
}

export async function updateProblem(
  slug: string,
  body: CreateProblemPayload,
) {
  return apiRequest<CreatedProblem>(`/problems/${slug}`, {
    method: "PATCH",
    body,
  });
}

export async function deleteProblem(slug: string) {
  return apiRequest<null>(`/problems/${slug}`, {
    method: "DELETE",
  });
}

export interface TestCaseRow {
  id: number;
  input: string;
  expected: string;
}

export interface TestCaseInput {
  input: string;
  expected: string;
}

export async function getHiddenTestCases(slug: string) {
  return apiRequest<TestCaseRow[]>(`/problems/${slug}/hidden-test-cases`);
}

export async function replaceHiddenTestCases(
  slug: string,
  testCases: TestCaseInput[],
) {
  return apiRequest<TestCaseRow[]>(`/problems/${slug}/hidden-test-cases`, {
    method: "PUT",
    body: { test_cases: testCases },
  });
}
