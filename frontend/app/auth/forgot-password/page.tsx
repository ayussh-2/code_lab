import { AuthShell } from "@/components/auth/auth-shell";
import { AuthPageContainer } from "@/components/auth/auth-page-container";
import { ForgotPasswordForm } from "./forgot-password-form";

export default function ForgotPasswordPage() {
    return (
        <AuthPageContainer>
            <AuthShell
                title="CODE_LAB"
                subtitle="Reset your password"
                footerText="Remembered it?"
                footerLinkText="Sign in"
                footerLinkHref="/auth/login"
            >
                <ForgotPasswordForm />
            </AuthShell>
        </AuthPageContainer>
    );
}
