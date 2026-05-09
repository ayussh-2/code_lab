import { TopNav } from "@/components/site/top-nav";
import { SessionBoot } from "@/components/auth/session-boot";

export default function DashboardLayout({
    children,
}: {
    children: React.ReactNode;
}) {
    return (
        <SessionBoot>
            <div>
                <TopNav />
                {children}
            </div>
        </SessionBoot>
    );
}
