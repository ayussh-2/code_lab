import type { ProblemDetail } from "@/lib/problems";
import { FileSystemTree } from "@webcontainer/api";

export const FRONTEND_PROBLEM: ProblemDetail = {
  id: 1001,
  title: "Live Character Counter",
  slug: "live-character-counter",
  difficulty: "easy",
  topics: ["React", "DOM", "Accessibility"],
  hints: [
    "Keep the input value in component state.",
    "The counter should update whenever the value changes.",
  ],
  details:
    "Build a live character counter for a text area. Show the current character count and prevent users from entering more than 280 characters.",
  examples: [
    {
      input: "Hello world",
      output: "11 / 280",
      explanation:
        "The counter reflects the number of characters currently entered.",
    },
  ],
  constraints: [
    "Render a labeled textarea.",
    "Update the count as the user types.",
    "Do not allow more than 280 characters.",
    "Use an accessible label and associate it with the textarea.",
  ],
  sample_test_cases: [
    {
      id: 1001,
      input: "Hello world",
      expected: "11",
    },
  ],
  acceptance_rate: 0,
  editorial_unlocked: false,
};

export const FRONTEND_PROJECT: FileSystemTree = {
  "package.json": {
    file: {
      contents: JSON.stringify(
        {
          name: "live-character-counter",
          private: true,
          version: "0.0.0",
          type: "module",
          scripts: {
            dev: "vite --host 0.0.0.0 --port 4173",
          },
          dependencies: {
            react: "^19.2.0",
            "react-dom": "^19.2.0",
          },
          devDependencies: {
            "@vitejs/plugin-react": "^5.1.0",
            vite: "^7.1.12",
          },
        },
        null,
        2,
      ),
    },
  },
  "index.html": {
    file: {
      contents: `<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Live Character Counter</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.jsx"></script>
  </body>
</html>`,
    },
  },
  "vite.config.js": {
    file: {
      contents: `import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({ plugins: [react()] });`,
    },
  },
  src: {
    directory: {
      "main.jsx": {
        file: {
          contents: `import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { App } from "./App";
import "./styles.css";

createRoot(document.getElementById("root")).render(
  <StrictMode>
    <App />
  </StrictMode>,
);`,
        },
      },
      "App.jsx": {
        file: {
          contents: `import { useState } from "react";

const MAX_LENGTH = 280;

export function App() {
  const [value, setValue] = useState("");

  return (
    <main className="shell">
      <p className="eyebrow">Frontend practice</p>
      <h1>Live Character Counter</h1>
      <p className="description">
        Write a short note and keep it within the character limit.
      </p>
      <label htmlFor="message">Your message</label>
      <textarea
        id="message"
        maxLength={MAX_LENGTH}
        value={value}
        onChange={(event) => setValue(event.target.value)}
        placeholder="Start typing..."
        rows={7}
      />
      <output aria-live="polite">{value.length} / {MAX_LENGTH}</output>
    </main>
  );
}`,
        },
      },
      "styles.css": {
        file: {
          contents: `:root {
  font-family: Inter, ui-sans-serif, system-ui, sans-serif;
  color: #f4f4f5;
  background: #09090b;
}

* { box-sizing: border-box; }
body { margin: 0; min-width: 320px; }
.shell {
  width: min(100% - 32px, 680px);
  margin: 10vh auto;
  padding: 32px;
  border: 1px solid #27272a;
  border-radius: 16px;
  background: #18181b;
}
.eyebrow {
  margin: 0 0 8px;
  color: #a1a1aa;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}
h1 { margin: 0; font-size: 32px; }
.description { color: #a1a1aa; }
label { display: block; margin: 24px 0 8px; font-weight: 600; }
textarea {
  display: block;
  width: 100%;
  resize: vertical;
  border: 1px solid #3f3f46;
  border-radius: 8px;
  padding: 12px;
  color: inherit;
  background: #09090b;
  font: inherit;
}
textarea:focus { outline: 2px solid #a1a1aa; outline-offset: 2px; }
output { display: block; margin-top: 8px; color: #a1a1aa; text-align: right; }`,
        },
      },
    },
  },
};
