import type { FileSystemTree } from "@webcontainer/api";

export const FRONTEND_QUESTION_PROJECT: FileSystemTree = {
    "package.json": {
        file: {
            contents: JSON.stringify(
                {
                    name: "frontend-question",
                    private: true,
                    version: "0.0.0",
                    type: "module",
                    scripts: { dev: "vite --host 0.0.0.0 --port 4173" },
                    dependencies: { react: "^19.2.0", "react-dom": "^19.2.0" },
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
    <title>Frontend Question Workspace</title>
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
import "./styles.css";
import { App } from "./App";

createRoot(document.getElementById("root")).render(
  <StrictMode>
    <App />
  </StrictMode>
);`,
                },
            },
            "App.jsx": {
                file: {
                    contents: `import { useState } from "react";

export function App() {
  const [name, setName] = useState("");

  return (
    <main className="shell">
      <h1>Frontend Question Sandbox</h1>
      <p>Edit this component and verify your solution in the preview panel.</p>
      <label htmlFor="name">Your answer:</label>
      <input
        id="name"
        value={name}
        onChange={(event) => setName(event.target.value)}
        placeholder="Type here..."
      />
      <div className="result">Current value: {name || "empty"}</div>
    </main>
  );
}`,
                },
            },
            "styles.css": {
                file: {
                    contents: `* { box-sizing: border-box; font-family: Inter, system-ui, sans-serif; }
body { margin: 0; background: #0b0b0b; color: #f3f4f6; }
.shell { max-width: 700px; margin: 40px auto; padding: 24px; border: 1px solid #262626; border-radius: 12px; background: #111111; }
h1 { margin: 0 0 8px; }
p { margin: 0 0 16px; color: #d4d4d8; }
label { display: block; margin-bottom: 8px; font-size: 14px; }
input { width: 100%; border: 1px solid #3f3f46; border-radius: 8px; padding: 10px 12px; background: #09090b; color: #f4f4f5; }
.result { margin-top: 12px; color: #e4e4e7; }`,
                },
            },
        },
    },
};
