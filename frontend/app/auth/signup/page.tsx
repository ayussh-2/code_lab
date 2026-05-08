import { AuthShell } from "@/components/auth/auth-shell";
import { AuthPageContainer } from "@/components/auth/auth-page-container";
import { SignupForm } from "./signup-form";

export default function SignupPage() {
  return (
    <AuthPageContainer>
      <AuthShell
        title="CODE_LAB"
        subtitle="Create your account"
        footerText="Already have an account?"
        footerLinkText="Sign in"
        footerLinkHref="/auth/login"
      >
        <SignupForm />
      </AuthShell>
    </AuthPageContainer>
  );
}
