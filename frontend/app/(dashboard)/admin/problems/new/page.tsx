"use client";

import { useRouter } from "next/navigation";
import { ProblemForm } from "@/components/admin/problem-form";
import { useAdminGuard } from "@/hooks/use-admin-guard";
import { createProblem } from "@/lib/problems";

export default function AdminNewProblemPage() {
    const router = useRouter();
    const allowed = useAdminGuard();

    if (!allowed) {
        return (
            <div className="flex min-h-[50vh] items-center justify-center bg-[#0e0e0e] text-sm text-zinc-500">
                Checking access...
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-[#0e0e0e] text-zinc-300">
            <main className="mx-auto max-w-[1100px] px-6 py-10">
                <header className="mb-10">
                    <p className="text-[11px] font-medium uppercase tracking-widest text-zinc-500">
                        Admin
                    </p>
                    <h1 className="mt-2 text-3xl font-semibold tracking-[-0.08em] text-white">
                        New problem
                    </h1>
                    <p className="mt-2 max-w-xl text-sm leading-7 text-zinc-500">
                        Statement uses Markdown with KaTeX. Backend stores raw
                        text; learners see the same render as the preview here.
                    </p>
                </header>

                <ProblemForm
                    mode="create"
                    submitLabel="Create problem"
                    submittingLabel="Creating..."
                    onSubmit={async (payload) => {
                        const result = await createProblem(payload);
                        router.push(`/problems/${result.data.slug}`);
                    }}
                />
            </main>
        </div>
    );
}
