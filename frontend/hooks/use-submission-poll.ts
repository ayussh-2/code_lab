"use client";

import { useEffect, useState } from "react";
import { ApiError } from "@/lib/api";
import {
    type Submission,
    getSubmission,
    isTerminal,
} from "@/lib/submissions";

interface UseSubmissionPollResult {
    submission: Submission | null;
    isPolling: boolean;
    errorMessage: string;
}

const DEFAULT_INTERVAL_MS = 1500;

export function useSubmissionPoll(
    id: number | null,
    intervalMs: number = DEFAULT_INTERVAL_MS,
): UseSubmissionPollResult {
    const [submission, setSubmission] = useState<Submission | null>(null);
    const [isPolling, setIsPolling] = useState(false);
    const [errorMessage, setErrorMessage] = useState("");
    const [errorSubmissionId, setErrorSubmissionId] = useState<number | null>(
        null,
    );

    useEffect(() => {
        if (id === null) return;

        let cancelled = false;
        let timeoutHandle: ReturnType<typeof setTimeout> | null = null;

        async function tick() {
            try {
                const result = await getSubmission(id as number);
                if (cancelled) return;

                setSubmission(result.data);
                setErrorMessage("");
                setErrorSubmissionId(null);

                if (isTerminal(result.data.status)) {
                    setIsPolling(false);
                    return;
                }

                timeoutHandle = setTimeout(tick, intervalMs);
            } catch (error) {
                if (cancelled) return;
                setErrorMessage(
                    error instanceof ApiError
                        ? error.message
                        : "Unable to fetch submission",
                );
                setErrorSubmissionId(id);
                setIsPolling(false);
            }
        }

        void tick();

        return () => {
            cancelled = true;
            if (timeoutHandle !== null) clearTimeout(timeoutHandle);
        };
    }, [id, intervalMs]);

    if (id === null) {
        return { submission: null, isPolling: false, errorMessage: "" };
    }

    const isCurrentSubmission = submission?.id === id;
    const currentErrorMessage =
        errorSubmissionId === id ? errorMessage : "";
    return {
        submission: isCurrentSubmission ? submission : null,
        isPolling:
            isPolling ||
            (!currentErrorMessage &&
                (!isCurrentSubmission || !isTerminal(submission.status))),
        errorMessage: currentErrorMessage,
    };
}
