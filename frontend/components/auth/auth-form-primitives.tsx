interface AuthErrorMessageProps {
  message: string;
}

interface AuthSubmitButtonProps {
  isLoading: boolean;
  idleText: string;
  loadingText: string;
}

export const AUTH_INPUT_CLASS_NAME =
  "w-full rounded-lg border border-[#1f1f1f] bg-black px-4 py-2 text-sm text-white placeholder:text-zinc-600 transition-colors focus:border-white focus:outline-none";

export function AuthErrorMessage({ message }: AuthErrorMessageProps) {
  if (!message) {
    return null;
  }

  return (
    <p className="rounded-lg bg-red-500/10 px-3 py-2 text-xs text-red-300">
      {message}
    </p>
  );
}

export function AuthSubmitButton({
  isLoading,
  idleText,
  loadingText,
}: AuthSubmitButtonProps) {
  return (
    <button
      className="mt-2 w-full rounded-lg bg-white px-4 py-3 text-xs font-semibold text-black transition-opacity hover:opacity-90 disabled:opacity-50"
      type="submit"
      disabled={isLoading}
    >
      {isLoading ? loadingText : idleText}
    </button>
  );
}

export function AuthSocialDivider() {
  return (
    <div className="my-6 flex items-center">
      <div className="h-px flex-1 bg-[#1f1f1f]" />
      <span className="px-3 text-[11px] uppercase tracking-[0.12em] text-zinc-600">
        Or continue with
      </span>
      <div className="h-px flex-1 bg-[#1f1f1f]" />
    </div>
  );
}
