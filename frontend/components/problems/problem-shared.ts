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
    "editorial",
    "submissions",
] as const;

export const EDITOR_LANGUAGES = [
    { label: "Python3", id: "python" },
    { label: "JavaScript", id: "javascript" },
    { label: "C++", id: "cpp" },
    { label: "Java", id: "java" },
] as const;

export type EditorLanguage = (typeof EDITOR_LANGUAGES)[number];

export const DEFAULT_EDITOR_CODE: Record<string, string> = {
    python: [
        "import sys",
        "",
        "def main():",
        "    data = sys.stdin.read().split()",
        "    # TODO: parse data and print the answer",
        "",
        "main()",
        "",
    ].join("\n"),
    javascript: [
        'const data = require("fs")',
        '    .readFileSync(0, "utf8")',
        "    .trim()",
        "    .split(/\\s+/);",
        "",
        "// TODO: parse data and print the answer using console.log",
        "",
    ].join("\n"),
    cpp: [
        "#include <bits/stdc++.h>",
        "using namespace std;",
        "",
        "int main() {",
        "    ios::sync_with_stdio(false);",
        "    cin.tie(nullptr);",
        "",
        "    // TODO: read input from cin and print the answer with cout",
        "",
        "    return 0;",
        "}",
        "",
    ].join("\n"),
    typescript: "function solve(): void {\n    \n};\n",
    c: "#include <stdio.h>\n\nint main() {\n    // TODO: read from stdin, print to stdout\n    return 0;\n}\n",
    java: `class Main {
    public static void main() {
        System.out.println("hello world!");
    }
    }`,
};
