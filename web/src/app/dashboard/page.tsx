"use client";

import { useAuth } from "@/components/AuthProvider";
import { subscribe, getApiKeys } from "@/lib/api";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";

export default function DashboardPage() {
  const { user, loading, signOut } = useAuth();
  const router = useRouter();
  const [apiKeys, setApiKeys] = useState<string[]>([]);
  const [subscribing, setSubscribing] = useState<string | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!loading && !user) {
      router.push("/login");
    }
  }, [user, loading, router]);

  useEffect(() => {
    if (user) {
      getApiKeys(user.uid)
        .then((data) => {
          if (Array.isArray(data)) {
            setApiKeys(data.map((k: { key: string }) => k.key));
          }
        })
        .catch(() => {});
    }
  }, [user]);

  const handleSubscribe = async (tier: "pro" | "premium") => {
    if (!user) return;
    setSubscribing(tier);
    setError("");
    try {
      const data = await subscribe(user.uid, user.email || "", tier);
      if (data.url) {
        window.location.href = data.url;
      }
    } catch (e) {
      setError("Subscription failed. Please try again.");
    } finally {
      setSubscribing(null);
    }
  };

  const handleSignOut = async () => {
    await signOut();
    router.push("/");
  };

  if (loading || !user) {
    return (
      <div className="flex flex-1 items-center justify-center">
        <div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-300 border-t-zinc-900" />
      </div>
    );
  }

  return (
    <div className="flex flex-col flex-1">
      {/* Dashboard Nav */}
      <header className="border-b border-zinc-200 bg-white dark:border-zinc-800 dark:bg-zinc-950">
        <div className="mx-auto flex h-14 max-w-4xl items-center justify-between px-4">
          <span className="text-lg font-bold tracking-tight">STX AI</span>
          <div className="flex items-center gap-4 text-sm">
            <span className="text-zinc-600 dark:text-zinc-400">
              {user.email}
            </span>
            <button
              onClick={handleSignOut}
              className="rounded-lg border border-zinc-300 px-3 py-1.5 text-sm font-medium hover:bg-zinc-50 dark:border-zinc-700 dark:hover:bg-zinc-800"
            >
              Sign Out
            </button>
          </div>
        </div>
      </header>

      <main className="mx-auto w-full max-w-4xl flex-1 px-4 py-8 space-y-8">
        {error && (
          <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-800 dark:bg-red-950 dark:text-red-400">
            {error}
          </div>
        )}

        {/* API Keys */}
        <section>
          <h2 className="text-xl font-semibold">API Keys</h2>
          <p className="mt-1 text-sm text-zinc-600 dark:text-zinc-400">
            Use these keys to authenticate your STX AI agent client.
          </p>
          <div className="mt-4 rounded-xl border border-zinc-200 dark:border-zinc-800">
            {apiKeys.length === 0 ? (
              <div className="px-4 py-8 text-center text-sm text-zinc-500">
                No API keys yet. Subscribe to a plan to get your first key.
              </div>
            ) : (
              <ul className="divide-y divide-zinc-200 dark:divide-zinc-800">
                {apiKeys.map((key, i) => (
                  <li
                    key={i}
                    className="flex items-center justify-between px-4 py-3"
                  >
                    <code className="rounded bg-zinc-100 px-2 py-1 font-mono text-sm dark:bg-zinc-800">
                      {key.slice(0, 12)}...
                    </code>
                    <button
                      onClick={() => navigator.clipboard.writeText(key)}
                      className="rounded px-2 py-1 text-xs font-medium text-zinc-500 hover:bg-zinc-100 dark:hover:bg-zinc-800"
                    >
                      Copy
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </section>

        {/* Subscription */}
        <section>
          <h2 className="text-xl font-semibold">Subscription</h2>
          <p className="mt-1 text-sm text-zinc-600 dark:text-zinc-400">
            Choose a plan to get API access and manage your subscription.
          </p>
          <div className="mt-4 grid gap-4 sm:grid-cols-2">
            <div className="rounded-xl border border-zinc-200 p-6 dark:border-zinc-800">
              <h3 className="font-semibold">Pro</h3>
              <p className="mt-1 text-3xl font-bold">
                HK$98<span className="text-sm font-normal text-zinc-500">/mo</span>
              </p>
              <ul className="mt-4 space-y-1 text-sm text-zinc-600 dark:text-zinc-400">
                <li>✓ 200 requests/day</li>
                <li>✓ Technical analysis</li>
                <li>✓ News aggregation</li>
                <li>✓ Watchlist</li>
              </ul>
              <button
                onClick={() => handleSubscribe("pro")}
                disabled={subscribing === "pro"}
                className="mt-4 w-full rounded-lg bg-zinc-900 px-4 py-2.5 text-sm font-semibold text-white hover:bg-zinc-800 disabled:opacity-50 dark:bg-zinc-100 dark:text-zinc-900 dark:hover:bg-zinc-200"
              >
                {subscribing === "pro" ? "Redirecting..." : "Subscribe Pro"}
              </button>
            </div>
            <div className="rounded-xl border border-blue-600 ring-1 ring-blue-600 p-6">
              <h3 className="font-semibold">Premium</h3>
              <p className="mt-1 text-3xl font-bold">
                HK$298<span className="text-sm font-normal text-zinc-500">/mo</span>
              </p>
              <ul className="mt-4 space-y-1 text-sm text-zinc-600 dark:text-zinc-400">
                <li>✓ 1,000 requests/day</li>
                <li>✓ Advanced scanning</li>
                <li>✓ Priority queue</li>
                <li>✓ Custom alerts</li>
              </ul>
              <button
                onClick={() => handleSubscribe("premium")}
                disabled={subscribing === "premium"}
                className="mt-4 w-full rounded-lg bg-blue-600 px-4 py-2.5 text-sm font-semibold text-white hover:bg-blue-700 disabled:opacity-50"
              >
                {subscribing === "premium" ? "Redirecting..." : "Subscribe Premium"}
              </button>
            </div>
          </div>
        </section>
      </main>
    </div>
  );
}
