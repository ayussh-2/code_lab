import type { ProblemDetail, ProblemListItem } from "@/lib/problems";

export const FRONTEND_QUESTIONS: ProblemDetail[] = [
    {
        id: 9001,
        title: "Build a Counter Component",
        slug: "build-a-counter-component",
        difficulty: "easy",
        topics: ["frontend", "react", "state"],
        details:
            "Implement a Counter component in React. The user should be able to increment, decrement, and reset the count. Display the current value clearly in the preview panel and keep the UI simple and readable.",
        hints: [
            "Use the useState hook to manage the count",
            "Provide three controls: Increment (+1), Decrement (-1), and Reset",
            "Prevent the count from going below 0",
            "Render the current count in a large, readable font",
        ],
        examples: [],
        constraints: [],
        sample_test_cases: [],
        acceptance_rate: 100,
        editorial_unlocked: true,
    },
];

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
