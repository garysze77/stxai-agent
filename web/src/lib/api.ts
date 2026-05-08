const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8000";

export async function subscribe(userId: string, email: string, tier: "pro" | "premium") {
  const res = await fetch(`${API_URL}/api/v1/subscribe`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ user_id: userId, email, tier }),
  });
  if (!res.ok) throw new Error("Subscription failed");
  return res.json();
}

export async function getApiKeys(userId: string) {
  const res = await fetch(`${API_URL}/api/v1/keys?user_id=${userId}`);
  if (!res.ok) return [];
  return res.json();
}
