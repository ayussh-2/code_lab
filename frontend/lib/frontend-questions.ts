import type { ProblemDetail, ProblemListItem } from "@/lib/problems";

export const FRONTEND_QUESTIONS: ProblemDetail[] = [];

export function getFrontendQuestionBySlug(
    slug: string,
): ProblemDetail | undefined {
    return FRONTEND_QUESTIONS.find((question) => question.slug === slug);
}

export function frontendQuestionsAsListItems(): ProblemListItem[] {
    return FRONTEND_QUESTIONS.map(
        ({ id, title, slug, difficulty, topics, acceptance_rate }) => ({
            id,
            title,
            slug,
            difficulty,
            topics,
            acceptance_rate,
        }),
    );
}
