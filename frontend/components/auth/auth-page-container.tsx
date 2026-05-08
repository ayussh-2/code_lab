import type { ReactNode } from "react";

interface AuthPageContainerProps {
  children: ReactNode;
}

export function AuthPageContainer({ children }: AuthPageContainerProps) {
  return (
    <div className="flex min-h-screen items-center justify-center bg-black px-4 py-8 text-zinc-100">
      {children}
    </div>
  );
}
