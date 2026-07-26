import type { NextConfig } from "next";

const nextConfig: NextConfig = {
    headers: async () => [
        {
            source: "/(.*)",
            headers: [
                { key: "Cross-Origin-Opener-Policy", value: "same-origin" },
                {
                    key: "Cross-Origin-Embedder-Policy",
                    value: "credentialless",
                },
            ],
        },
    ],
};

export default nextConfig;

import('@opennextjs/cloudflare').then(m => m.initOpenNextCloudflareForDev());
