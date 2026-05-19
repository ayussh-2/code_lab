"use client";

import { useParams } from "next/navigation";
import { PublicProfileView } from "@/components/profile/public-profile-view";

export default function PublicUserProfilePage() {
  const params = useParams<{ username: string }>();
  const username = params.username ?? "";

  return (
    <div className="min-h-screen bg-[#0e0e0e] text-zinc-300">
      <PublicProfileView username={username} />
    </div>
  );
}
