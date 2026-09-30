import type { FileSystemTree } from "@webcontainer/api";

export interface FrontendQuestion {
  title: string;
  description: string;
  requirements: string[];
}

export type ActivePanel = "question" | "files";

export interface WebContProps {
  files: FileSystemTree;
  question: FrontendQuestion;
}
