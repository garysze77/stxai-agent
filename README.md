# STX AI - 自主金融 AI Agent

開源金融 AI Agent，專注美股及港股市場分析。用戶自行安裝 Agent 客戶端，AI 分析能力由 STX AI Cloud API 提供，需 Subscription 獲取 API Key。

## 架構

- `cloud/` — STX AI Cloud API (付費 SaaS，部署於 Railway)
- `web/` — Landing Page + User Dashboard (Vercel + Firebase Auth)
- `agent/` — STX AI Agent 客戶端 (開源，pip install)

## 快速開始

### Cloud API

```bash
cd cloud
cp .env.example .env
# 編輯 .env 填上必要設定
pip install -e .
uvicorn src.main:app --reload
```

### Agent Client (Phase 6)

```bash
pip install stxai-agent
stxai setup
stxai start
```

## 技術棧

| 組件 | 技術 |
|------|------|
| Cloud API | Python FastAPI (Railway) |
| LLM | Puter API (主) + MiniMax (後備) |
| Auth + DB | Firebase (Auth + Firestore) |
| 支付 | Stripe |
| Landing | Next.js (Vercel) |
| Agent | Python (開源) |
