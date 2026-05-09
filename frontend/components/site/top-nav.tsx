import Link from "next/link";
import { ProfileMenu } from "@/components/site/profile-menu";

export function TopNav() {
    return (
        <header className="sticky top-0 z-50 border-b border-white/[0.08] bg-[#141414] w-full">
            <div className="mx-auto flex items-center justify-between px-6 py-2.5 w-full">
                <div className="flex items-center justify-between w-full gap-6">
                    <Link
                        href="/problems"
                        className="text-base font-semibold tracking-[-0.03em] text-white"
                    >
                        CODE_LAB
                    </Link>
                    <nav className="flex items-center gap-4">
                        <Link
                            href="/problems"
                            className="text-sm text-zinc-400 transition-colors hover:text-white"
                        >
                            Problems
                        </Link>
                        <ProfileMenu />
                    </nav>
                </div>
            </div>
        </header>
    );
}
