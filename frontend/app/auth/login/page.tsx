import { Suspense } from "react";
import { AuthShell } from "@/components/auth/auth-shell";
import { AuthPageContainer } from "@/components/auth/auth-page-container";
import { LoginForm } from "./login-form";

export default function LoginPage() {
  return (
    <AuthPageContainer>
      <AuthShell
        title="CODE_LAB"
        subtitle="Sign in to your account"
        footerText="Don't have an account?"
        footerLinkText="Sign up"
        footerLinkHref="/auth/signup"
      >
        <Suspense fallback={null}>
          <LoginForm />
        </Suspense>
      </AuthShell>
    </AuthPageContainer>
  );
}
