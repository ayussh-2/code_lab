import { Suspense } from "react";
import { AuthShell } from "@/components/auth/auth-shell";
import { AuthPageContainer } from "@/components/auth/auth-page-container";
import { VerifyEmailForm } from "./verify-email-form";

export default function VerifyEmailPage() {
    return (
        <AuthPageContainer>
            <AuthShell
                title="CODE_LAB"
                subtitle="Verify your email"
                footerText="Wrong account?"
                footerLinkText="Sign up again"
                footerLinkHref="/auth/signup"
            >
                <Suspense fallback={null}>
                    <VerifyEmailForm />
                </Suspense>
            </AuthShell>
        </AuthPageContainer>
    );
}
