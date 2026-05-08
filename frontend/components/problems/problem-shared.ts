export const FRONTEND_PROBLEM_TOPIC = "frontend";

export function isFrontendProblem(topics: string[]): boolean {
    return topics.some(
        (topic) => topic.toLowerCase() === FRONTEND_PROBLEM_TOPIC,
    );
}

export function getProblemHref(problem: {
    slug: string;
    topics: string[];
}): string {
    return isFrontendProblem(problem.topics)
        ? `/problems/frontend/${problem.slug}`
        : `/problems/${problem.slug}`;
}

export const PROBLEM_DIFFICULTY_TEXT_CLASS: Record<string, string> = {
    easy: "text-[#1cbf73]",
    medium: "text-[#e8a24a]",
    hard: "text-[#ef4743]",
};

export const PROBLEM_DIFFICULTY_BADGE_CLASS: Record<string, string> = {
    easy: "bg-[#1cbf73]/10 text-[#1cbf73]",
    medium: "bg-[#e8a24a]/10 text-[#e8a24a]",
    hard: "bg-[#ef4743]/10 text-[#ef4743]",
};

export const PROBLEM_DETAIL_TABS = [
    "description",
    // "editorial",
    // "solutions",
    "submissions",
] as const;

export const EDITOR_LANGUAGES = [
    { label: "Python3", id: "python" },
    { label: "JavaScript", id: "javascript" },
    { label: "TypeScript", id: "typescript" },
    { label: "C++", id: "cpp" },
    { label: "C", id: "c" },
    { label: "Java", id: "java" },
    { label: "Go", id: "go" },
    { label: "Rust", id: "rust" },
    { label: "C#", id: "csharp" },
    { label: "Swift", id: "swift" },
    { label: "Kotlin", id: "kotlin" },
    { label: "Ruby", id: "ruby" },
] as const;

export type EditorLanguage = (typeof EDITOR_LANGUAGES)[number];

export const DEFAULT_EDITOR_CODE: Record<string, string> = {
    python: "class Solution:\n    def solve(self):\n        pass\n",
    javascript:
        "/**\n * @return {void}\n */\nvar solve = function() {\n    \n};\n",
    typescript: "function solve(): void {\n    \n};\n",
    cpp: "#include <bits/stdc++.h>\nusing namespace std;\n\nclass Solution {\npublic:\n    void solve() {\n        \n    }\n};\n",
    c: "#include <stdio.h>\n\nvoid solve() {\n    \n}\n",
    java: "class Solution {\n    public void solve() {\n        \n    }\n}\n",
    go: "package main\n\nfunc solve() {\n    \n}\n",
    rust: "impl Solution {\n    pub fn solve() {\n        \n    }\n}\n",
    csharp: "public class Solution {\n    public void Solve() {\n        \n    }\n}\n",
    swift: "class Solution {\n    func solve() {\n        \n    }\n}\n",
    kotlin: "class Solution {\n    fun solve() {\n        \n    }\n}\n",
    ruby: "def solve\n    \nend\n",
};
