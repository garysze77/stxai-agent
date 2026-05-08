package bot

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/garysze77/stxai/agent/internal/client"
	"github.com/garysze77/stxai/agent/internal/store"

	tele "gopkg.in/telebot.v3"
)

type Bot struct {
	tg     *tele.Bot
	client *client.Client
	store  *store.Store
}

func New(token string, c *client.Client, s *store.Store) (*Bot, error) {
	pref := tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	tg, err := tele.NewBot(pref)
	if err != nil {
		return nil, fmt.Errorf("telegram bot init: %w", err)
	}

	return &Bot{tg: tg, client: c, store: s}, nil
}

func (b *Bot) Start(ctx context.Context) error {
	b.registerHandlers()
	log.Printf("🤖 STX AI Bot starting... t.me/%s", b.tg.Me.Username)
	b.tg.Start()
	<-ctx.Done()
	b.tg.Stop()
	return nil
}

func (b *Bot) registerHandlers() {
	b.tg.Handle("/start", b.handleStart)
	b.tg.Handle("/help", b.handleHelp)
	b.tg.Handle("/analyze", b.handleAnalyze)
	b.tg.Handle("/clear", b.handleClear)
	b.tg.Handle(tele.OnText, b.handleText)
}

func (b *Bot) handleStart(c tele.Context) error {
	msg := "🚀 STX AI Agent 已啟動！\n\n" +
		"/analyze <ticker> — 股票分析 (e.g. /analyze AAPL)\n" +
		"/clear — 清除對話記錄\n" +
		"/help — 幫助\n\n" +
		"直接 send message 就可以同 AI 對話分析股票。"
	return c.Send(msg)
}

func (b *Bot) handleHelp(c tele.Context) error {
	msg := "📖 STX AI 使用方法：\n\n" +
		"• 直接問股票：AAPL 現價幾多？\n" +
		"• 技術分析：TSLA RSI 同 MACD 點睇？\n" +
		"• 市場掃描：港股今日邊隻最活躍？\n" +
		"• 新聞：NVDA 最新消息？\n\n" +
		"支援美股＋港股（港股加 .HK：0700.HK）"
	return c.Send(msg)
}

func (b *Bot) handleAnalyze(c tele.Context) error {
	ticker := strings.TrimSpace(c.Message().Payload)
	if ticker == "" {
		return c.Send("Usage: /analyze <ticker>\nExample: /analyze AAPL")
	}

	c.Send("📊 Analyzing " + ticker + "...")

	ar, err := b.client.Analyze(ticker)
	if err != nil {
		return c.Send("❌ " + err.Error())
	}

	msg := fmt.Sprintf(
		"📈 *%s*  $%.2f (%.2f%%)\n\n"+
			"RSI: %.1f | MACD: %.2f\n"+
			"MA50: $%.2f | MA200: $%.2f\n"+
			"Bollinger: $%.2f — $%.2f\n\n"+
			"📰 %s",
		ar.Ticker, ar.Price, ar.Change,
		ar.RSI, ar.MACD,
		ar.MA50, ar.MA200,
		ar.BollLower, ar.BollUpper,
		ar.Summary,
	)
	return c.Send(msg, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
}

func (b *Bot) handleClear(c tele.Context) error {
	sessionID := fmt.Sprintf("tg:%d", c.Sender().ID)
	b.store.DeleteOldMessages(time.Now()) // clean all for this session
	return c.Send("✅ 對話記錄已清除。Session: " + sessionID)
}

func (b *Bot) handleText(c tele.Context) error {
	message := c.Text()
	sessionID := fmt.Sprintf("tg:%d", c.Sender().ID)

	c.Notify(tele.Typing)

	// Save user message
	b.store.SaveMessage(sessionID, "user", message)

	resp, err := b.client.Chat(message, sessionID)
	if err != nil {
		return c.Send("❌ " + err.Error())
	}

	// Save assistant message
	b.store.SaveMessage(sessionID, "assistant", resp.Reply)

	if len(resp.Reply) > 4000 {
		return b.sendLongMessage(c, resp.Reply)
	}
	return c.Send(resp.Reply, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
}

func (b *Bot) sendLongMessage(c tele.Context, text string) error {
	for i := 0; i < len(text); i += 4000 {
		end := i + 4000
		if end > len(text) {
			end = len(text)
		}
		if err := c.Send(text[i:end]); err != nil {
			return err
		}
	}
	return nil
}
