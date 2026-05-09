"use client";

import { useAuth } from "@/components/AuthProvider";
import Link from "next/link";
import Image from "next/image";

const FEATURES = [
  {
    title: "AI-Powered Analysis",
    desc: "Deep market insights powered by advanced language models. Get plain-English explanations of technical indicators, trends, and risks for any US or HK stock.",
    icon: "🤖",
  },
  {
    title: "Real-Time Market Data",
    desc: "Live prices, historical data, and market scanning across US and Hong Kong exchanges. yfinance + akshare integration for comprehensive coverage.",
    icon: "📊",
  },
  {
    title: "Technical Indicators",
    desc: "Built-in RSI, MACD, Bollinger Bands, moving averages, and more. No manual chart reading — the AI interprets everything for you.",
    icon: "📈",
  },
  {
    title: "Self-Hosted Agent",
    desc: "Open-source Go client — single binary, zero dependencies. Install in seconds. CLI + Telegram bot included. You own your data.",
    icon: "💻",
  },
  {
    title: "Subscription API Access",
    desc: "Pay only for what you use. Simple tiered plans with daily quotas. Free tier available to try before you commit.",
    icon: "🔑",
  },
  {
    title: "Multi-Market Coverage",
    desc: "US stocks (NYSE, NASDAQ) and Hong Kong stocks (HKEX). Expand to forex and crypto on the roadmap.",
    icon: "🌏",
  },
];

const TIERS = [
  {
    name: "Free",
    price: "$0",
    period: "forever",
    quota: "10 req/day",
    features: ["Basic stock quotes", "Simple Q&A", "Single stock analysis"],
    cta: "Get Started",
    href: "/login",
    highlight: false,
  },
  {
    name: "Pro",
    price: "HK$98",
    period: "/month",
    quota: "200 req/day",
    features: [
      "Everything in Free",
      "Technical analysis",
      "News aggregation",
      "Watchlist",
      "Email support",
    ],
    cta: "Subscribe",
    href: "/login",
    highlight: true,
  },
  {
    name: "Premium",
    price: "HK$298",
    period: "/month",
    quota: "1,000 req/day",
    features: [
      "Everything in Pro",
      "Advanced market scanning",
      "Priority queue",
      "Custom alerts",
      "Priority support",
    ],
    cta: "Subscribe",
    href: "/login",
    highlight: false,
  },
];

export default function LandingPage() {
  const { user } = useAuth();

  return (
    <div className="flex flex-col flex-1">
      {/* Nav */}
      <header className="sticky top-0 z-50 border-b border-zinc-200 bg-white/80 backdrop-blur dark:border-zinc-800 dark:bg-zinc-950/80">
        <div className="mx-auto flex h-14 max-w-6xl items-center justify-between px-4">
          <Link href="/" className="flex items-center gap-2">
            <Image src="/stxai-logo.png" alt="STX AI" width={32} height={32} className="rounded" />
            <span className="text-lg font-bold tracking-tight">STX AI</span>
          </Link>
          <nav className="flex items-center gap-4 text-sm font-medium">
            <a href="#features" className="text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100">
              Features
            </a>
            <a href="#pricing" className="text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100">
              Pricing
            </a>
            {user ? (
              <Link
                href="/dashboard"
                className="rounded-lg bg-zinc-900 px-4 py-2 text-white hover:bg-zinc-800 dark:bg-zinc-100 dark:text-zinc-900 dark:hover:bg-zinc-200"
              >
                Dashboard
              </Link>
            ) : (
              <Link
                href="/login"
                className="rounded-lg bg-zinc-900 px-4 py-2 text-white hover:bg-zinc-800 dark:bg-zinc-100 dark:text-zinc-900 dark:hover:bg-zinc-200"
              >
                Sign In
              </Link>
            )}
          </nav>
        </div>
      </header>

      {/* Hero */}
      <section className="px-4 py-24 text-center">
        <h1 className="mx-auto max-w-3xl text-5xl font-bold tracking-tight sm:text-6xl">
          Autonomous Financial AI Agent for{" "}
          <span className="text-blue-600">US & HK Markets</span>
        </h1>
        <p className="mx-auto mt-6 max-w-xl text-lg text-zinc-600 dark:text-zinc-400">
          Open-source, self-hosted AI agent that analyzes stocks, runs technical
          indicators, and delivers plain-English insights. Subscribe for API access.
        </p>
        <div className="mt-8 flex items-center justify-center gap-4">
          <Link
            href="/login"
            className="rounded-xl bg-zinc-900 px-6 py-3 text-sm font-semibold text-white hover:bg-zinc-800 dark:bg-zinc-100 dark:text-zinc-900 dark:hover:bg-zinc-200"
          >
            Get Started Free
          </Link>
          <a
            href="#features"
            className="rounded-xl border border-zinc-300 px-6 py-3 text-sm font-semibold text-zinc-700 hover:bg-zinc-50 dark:border-zinc-700 dark:text-zinc-300 dark:hover:bg-zinc-900"
          >
            Learn More
          </a>
        </div>
      </section>

      {/* Install */}
      <section className="px-4 py-20">
        <div className="mx-auto max-w-3xl text-center">
          <h2 className="text-3xl font-bold tracking-tight">
            Install in Seconds
          </h2>
          <p className="mt-4 text-zinc-600 dark:text-zinc-400">
            Single binary, zero dependencies. Works on macOS, Linux, and Windows.
          </p>
          <div className="mt-6 rounded-xl border border-zinc-200 bg-zinc-900 p-4 dark:border-zinc-700">
            <div className="flex items-center justify-between gap-3">
              <code className="text-sm text-green-400 overflow-x-auto whitespace-nowrap">
                curl -fsSL https://raw.githubusercontent.com/garysze77/stxai-agent/main/install.sh | sh
              </code>
              <button
                onClick={() => navigator.clipboard.writeText("curl -fsSL https://raw.githubusercontent.com/garysze77/stxai-agent/main/install.sh | sh")}
                className="shrink-0 rounded-lg border border-zinc-700 px-3 py-1.5 text-xs font-medium text-zinc-300 hover:bg-zinc-800"
              >
                Copy
              </button>
            </div>
          </div>
          <div className="mt-6 grid gap-4 sm:grid-cols-3 text-sm text-zinc-600 dark:text-zinc-400">
            <div className="rounded-xl border border-zinc-200 p-4 dark:border-zinc-800">
              <span className="text-lg font-bold text-zinc-900 dark:text-zinc-100">1</span>
              <p className="mt-1">Run the install command above</p>
            </div>
            <div className="rounded-xl border border-zinc-200 p-4 dark:border-zinc-800">
              <span className="text-lg font-bold text-zinc-900 dark:text-zinc-100">2</span>
              <p className="mt-1"><code className="rounded bg-zinc-100 px-1 dark:bg-zinc-800">stxai setup</code> — enter your API key</p>
            </div>
            <div className="rounded-xl border border-zinc-200 p-4 dark:border-zinc-800">
              <span className="text-lg font-bold text-zinc-900 dark:text-zinc-100">3</span>
              <p className="mt-1"><code className="rounded bg-zinc-100 px-1 dark:bg-zinc-800">stxai chat</code> — start analyzing</p>
            </div>
          </div>
          <p className="mt-4 text-xs text-zinc-400">
            Or{" "}
            <a href="https://github.com/garysze77/stxai-agent" target="_blank" rel="noopener noreferrer" className="underline hover:text-zinc-600">
              download from GitHub Releases
            </a>
            {" "}·{" "}
            <a href="https://github.com/garysze77/stxai-agent" target="_blank" rel="noopener noreferrer" className="underline hover:text-zinc-600">
              build from source
            </a>
          </p>
        </div>
      </section>

      {/* Features */}
      <section id="features" className="px-4 py-20 bg-zinc-50 dark:bg-zinc-950">
        <div className="mx-auto max-w-6xl">
          <h2 className="text-center text-3xl font-bold tracking-tight">
            Everything You Need to Analyze Markets
          </h2>
          <p className="mx-auto mt-4 max-w-lg text-center text-zinc-600 dark:text-zinc-400">
            From raw data to actionable insights — all in one agent.
          </p>
          <div className="mt-12 grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
            {FEATURES.map((f) => (
              <div
                key={f.title}
                className="rounded-2xl border border-zinc-200 bg-white p-6 dark:border-zinc-800 dark:bg-zinc-900"
              >
                <div className="text-2xl">{f.icon}</div>
                <h3 className="mt-3 text-lg font-semibold">{f.title}</h3>
                <p className="mt-2 text-sm text-zinc-600 dark:text-zinc-400">
                  {f.desc}
                </p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Pricing */}
      <section id="pricing" className="px-4 py-20">
        <div className="mx-auto max-w-6xl">
          <h2 className="text-center text-3xl font-bold tracking-tight">
            Simple, Transparent Pricing
          </h2>
          <p className="mx-auto mt-4 max-w-lg text-center text-zinc-600 dark:text-zinc-400">
            Start free, upgrade when you need more. Cancel anytime.
          </p>
          <div className="mt-12 grid gap-6 lg:grid-cols-3">
            {TIERS.map((t) => (
              <div
                key={t.name}
                className={`relative rounded-2xl border p-6 ${
                  t.highlight
                    ? "border-blue-600 ring-1 ring-blue-600 bg-white dark:bg-zinc-900"
                    : "border-zinc-200 bg-white dark:border-zinc-800 dark:bg-zinc-900"
                }`}
              >
                {t.highlight && (
                  <span className="absolute -top-3 left-1/2 -translate-x-1/2 rounded-full bg-blue-600 px-3 py-0.5 text-xs font-semibold text-white">
                    Most Popular
                  </span>
                )}
                <h3 className="text-lg font-semibold">{t.name}</h3>
                <div className="mt-2">
                  <span className="text-4xl font-bold">{t.price}</span>
                  <span className="text-sm text-zinc-500"> {t.period}</span>
                </div>
                <p className="mt-1 text-sm font-medium text-zinc-500">
                  {t.quota}
                </p>
                <ul className="mt-6 space-y-2 text-sm text-zinc-600 dark:text-zinc-400">
                  {t.features.map((f) => (
                    <li key={f} className="flex items-center gap-2">
                      <span className="text-green-600">✓</span> {f}
                    </li>
                  ))}
                </ul>
                <Link
                  href={t.href}
                  className={`mt-6 block w-full rounded-xl px-4 py-2.5 text-center text-sm font-semibold ${
                    t.highlight
                      ? "bg-blue-600 text-white hover:bg-blue-700"
                      : "bg-zinc-900 text-white hover:bg-zinc-800 dark:bg-zinc-100 dark:text-zinc-900 dark:hover:bg-zinc-200"
                  }`}
                >
                  {t.cta}
                </Link>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="mt-auto border-t border-zinc-200 bg-zinc-50 px-4 py-8 dark:border-zinc-800 dark:bg-zinc-950">
        <div className="mx-auto flex max-w-6xl items-center justify-between text-sm text-zinc-500">
          <span>STX AI — Open Source Financial Agent</span>
          <a
            href="https://github.com/garysze77/stxai-agent"
            target="_blank"
            rel="noopener noreferrer"
            className="hover:text-zinc-700 dark:hover:text-zinc-300"
          >
            GitHub
          </a>
        </div>
      </footer>
    </div>
  );
}
