"use client";

import { ReactNode, useEffect, useState } from "react";
import { ensureSession } from "@/lib/auth";

interface SessionBootProps {
    children: ReactNode;
}

export function SessionBoot({ children }: SessionBootProps) {
    const [ready, setReady] = useState(false);

    useEffect(() => {
        let cancelled = false;
        void ensureSession().finally(() => {
            if (!cancelled) setReady(true);
        });
        return () => {
            cancelled = true;
        };
    }, []);

    if (!ready) {
        return (
            <div className="flex min-h-screen items-center justify-center bg-[#0e0e0e] text-xs text-zinc-500">
                Loading...
            </div>
        );
    }

    return <>{children}</>;
}
