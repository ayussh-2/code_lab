interface LoadingStateProps {
  message: string;
  className?: string;
}

interface ErrorStateProps {
  message: string;
  className?: string;
}

export function LoadingState({ message, className = "" }: LoadingStateProps) {
  return <p className={`text-sm text-zinc-500 ${className}`.trim()}>{message}</p>;
}

export function ErrorState({ message, className = "" }: ErrorStateProps) {
  return <p className={`text-sm text-red-400 ${className}`.trim()}>{message}</p>;
}
