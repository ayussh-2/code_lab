import { TopNav } from "@/components/site/top-nav";
import { AuthProvider } from "@/components/auth/auth-context";

export default function PublicProfileLayout({
    children,
}: {
    children: React.ReactNode;
}) {
    return (
        <AuthProvider>
            <div>
                <TopNav />
                {children}
            </div>
        </AuthProvider>
    );
}
