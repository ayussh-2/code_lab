import { NextRequest, NextResponse } from "next/server";

const ACCESS_COOKIE =
    process.env.NEXT_PUBLIC_ACCESS_COOKIE_NAME ?? "access_token";

export function middleware(request: NextRequest) {
    const hasCookie = Boolean(request.cookies.get(ACCESS_COOKIE)?.value);

    if (!hasCookie) {
        const loginUrl = new URL("/auth/login", request.url);
        loginUrl.searchParams.set("from", request.nextUrl.pathname);
        return NextResponse.redirect(loginUrl);
    }

    return NextResponse.next();
}

export const config = {
    matcher: ["/profile/:path*", "/admin/:path*"],
};
