import { TopNav } from "@/components/site/top-nav";

export default function DashboardLayout({
    children,
}: {
    children: React.ReactNode;
}) {
    return (
        <div>
            <TopNav />
            {children}
        </div>
    );
}
