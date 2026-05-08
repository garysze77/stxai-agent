package bot

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/garysze77/stxai-agent/internal/client"
	"github.com/garysze77/stxai-agent/internal/store"

	tele "gopkg.in/telebot.v3"
)

// Keyboard button labels (also used as callback text)
const (
	btnAnalyze  = "📈 Analyze"
	btnSignal   = "⚡ Signal"
	btnNews     = "📰 News"
	btnScanUS   = "📊 US Scan"
	btnScanHK   = "🇭🇰 HK Scan"
	btnClear    = "🗑 Clear"
	btnHelp     = "ℹ️ Help"
	btnCancel   = "❌ Cancel"
)

// Mode states for keyboard flow
const (
	modeNone    = ""
	modeAnalyze = "analyze"
	modeSignal  = "signal"
	modeNews    = "news"
)

type Bot struct {
	tg     *tele.Bot
	client *client.Client
	store  *store.Store

	// Track which mode each user is in (userID → mode)
	mu    sync.Mutex
	modes map[int64]string
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

	return &Bot{
		tg:     tg,
		client: c,
		store:  s,
		modes:  make(map[int64]string),
	}, nil
}

// ── Mode management ──

func (b *Bot) setMode(userID int64, mode string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if mode == modeNone {
		delete(b.modes, userID)
	} else {
		b.modes[userID] = mode
	}
}

func (b *Bot) getMode(userID int64) string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.modes[userID]
}

// ── Keyboard ──

func mainKeyboard() *tele.ReplyMarkup {
	m := &tele.ReplyMarkup{ResizeKeyboard: true}
	m.ReplyKeyboard = [][]tele.ReplyButton{
		{{Text: btnAnalyze}, {Text: btnSignal}, {Text: btnNews}},
		{{Text: btnScanUS}, {Text: btnScanHK}},
		{{Text: btnClear}, {Text: btnHelp}},
	}
	return m
}

func cancelKeyboard() *tele.ReplyMarkup {
	m := &tele.ReplyMarkup{ResizeKeyboard: true}
	m.ReplyKeyboard = [][]tele.ReplyButton{
		{{Text: btnCancel}},
	}
	return m
}

// ── Startup ──

func (b *Bot) Start(ctx context.Context) error {
	b.registerHandlers()
	b.setCommands()
	log.Printf("🤖 STX AI Bot starting... t.me/%s", b.tg.Me.Username)
	b.tg.Start()
	<-ctx.Done()
	b.tg.Stop()
	return nil
}

func (b *Bot) setCommands() {
	cmds := []tele.Command{
		{Text: "analyze", Description: "Deep multi-agent stock analysis"},
		{Text: "scan", Description: "Scan market for top movers"},
		{Text: "news", Description: "Latest news for a stock"},
		{Text: "signal", Description: "Trading signal without deep re-analysis"},
		{Text: "clear", Description: "Clear conversation history"},
		{Text: "help", Description: "Show usage guide"},
	}
	if err := b.tg.SetCommands(cmds); err != nil {
		log.Printf("⚠️ setCommands failed: %v", err)
	}
}

func (b *Bot) registerHandlers() {
	b.tg.Handle("/start", b.handleStart)
	b.tg.Handle("/help", b.handleHelp)
	b.tg.Handle("/analyze", b.handleAnalyze)
	b.tg.Handle("/scan", b.handleScan)
	b.tg.Handle("/news", b.handleNews)
	b.tg.Handle("/signal", b.handleSignal)
	b.tg.Handle("/clear", b.handleClear)
	b.tg.Handle(tele.OnText, b.handleText)
}

// ── Slash command handlers ──

func (b *Bot) handleStart(c tele.Context) error {
	b.setMode(c.Sender().ID, modeNone)

	msg := "🚀 *STX AI Agent* — Autonomous Financial AI\n\n" +
		"Tap a button below or type a stock\\-related question\\!\n" +
		"Supports US \\(NYSE, NASDAQ\\) \\& Hong Kong \\(HKEX\\) stocks\\."

	return c.Send(msg, &tele.SendOptions{
		ParseMode: tele.ModeMarkdownV2,
		ReplyMarkup: mainKeyboard(),
	})
}

func (b *Bot) handleHelp(c tele.Context) error {
	b.setMode(c.Sender().ID, modeNone)

	msg := "📖 *STX AI Usage*\n\n" +
		"• Type a ticker: `AAPL`, `$TSLA`, `0700\\.HK`\n" +
		"• Ask for analysis: `NVDA technicals?`\n" +
		"• Use the keyboard below for quick actions\n\n" +
		"*Buttons:*\n" +
		"📈 Analyze — Deep multi\\-agent report\n" +
		"⚡ Signal — Quick trading signal\n" +
		"📰 News — Latest headlines\n" +
		"📊 US/HK Scan — Market movers\n\n" +
		"Supports US & Hong Kong stocks\\."

	return c.Send(msg, &tele.SendOptions{
		ParseMode:   tele.ModeMarkdownV2,
		ReplyMarkup: mainKeyboard(),
	})
}

func (b *Bot) handleAnalyze(c tele.Context) error {
	ticker := strings.TrimSpace(c.Message().Payload)
	if ticker == "" {
		return b.promptTicker(c, modeAnalyze, "🔬 *Deep Analysis*\n\nEnter a stock code\\(e\\.g\\. AAPL, 0700\\):")
	}
	return b.runAnalyze(c, ticker)
}

func (b *Bot) handleSignal(c tele.Context) error {
	ticker := strings.TrimSpace(c.Message().Payload)
	if ticker == "" {
		return b.promptTicker(c, modeSignal, "⚡ *Quick Signal*\n\nEnter a stock code\\(e\\.g\\. AAPL, 0700\\):")
	}
	return b.runSignal(c, ticker)
}

func (b *Bot) handleNews(c tele.Context) error {
	ticker := strings.TrimSpace(c.Message().Payload)
	if ticker == "" {
		return b.promptTicker(c, modeNews, "📰 *News*\n\nEnter a stock code\\(e\\.g\\. AAPL, 0700\\):")
	}
	return b.runNews(c, ticker)
}

func (b *Bot) handleScan(c tele.Context) error {
	market := strings.TrimSpace(c.Message().Payload)
	if market == "" {
		market = "us"
	}
	if market != "us" && market != "hk" {
		return c.Send("Usage: `/scan us` or `/scan hk`",
			&tele.SendOptions{ParseMode: tele.ModeMarkdownV2, ReplyMarkup: mainKeyboard()})
	}
	return b.runScan(c, market)
}

func (b *Bot) handleClear(c tele.Context) error {
	b.setMode(c.Sender().ID, modeNone)
	b.store.DeleteOldMessages(time.Now())
	return c.Send("✅ History cleared\\.", &tele.SendOptions{
		ParseMode:   tele.ModeMarkdownV2,
		ReplyMarkup: mainKeyboard(),
	})
}

// ── Keyboard button / text handler ──

func (b *Bot) handleText(c tele.Context) error {
	text := c.Text()
	userID := c.Sender().ID

	// 1. Check if it's a keyboard button press
	switch text {
	case btnAnalyze:
		return b.promptTicker(c, modeAnalyze, "🔬 *Deep Analysis*\n\nEnter a stock code\\(e\\.g\\. AAPL, 0700\\):")
	case btnSignal:
		return b.promptTicker(c, modeSignal, "⚡ *Quick Signal*\n\nEnter a stock code\\(e\\.g\\. AAPL, 0700\\):")
	case btnNews:
		return b.promptTicker(c, modeNews, "📰 *News*\n\nEnter a stock code\\(e\\.g\\. AAPL, 0700\\):")
	case btnScanUS:
		return b.runScan(c, "us")
	case btnScanHK:
		return b.runScan(c, "hk")
	case btnClear:
		return b.handleClear(c)
	case btnHelp:
		return b.handleHelp(c)
	case btnCancel:
		b.setMode(userID, modeNone)
		return c.Send("✅ Cancelled\\. Back to normal chat\\.", &tele.SendOptions{
			ParseMode:   tele.ModeMarkdownV2,
			ReplyMarkup: mainKeyboard(),
		})
	}

	// 2. Check if user is in a mode (waiting for ticker input)
	mode := b.getMode(userID)
	switch mode {
	case modeAnalyze:
		return b.runAnalyze(c, text)
	case modeSignal:
		return b.runSignal(c, text)
	case modeNews:
		return b.runNews(c, text)
	}

	// 3. Normal chat — send to cloud API
	return b.runChat(c, text)
}

// ── Prompt for ticker ──

func (b *Bot) promptTicker(c tele.Context, mode, msg string) error {
	b.setMode(c.Sender().ID, mode)
	return c.Send(msg, &tele.SendOptions{
		ParseMode:   tele.ModeMarkdownV2,
		ReplyMarkup: cancelKeyboard(),
	})
}

// ── Action runners ──

func (b *Bot) runAnalyze(c tele.Context, ticker string) error {
	b.setMode(c.Sender().ID, modeNone)
	ticker = cleanTicker(ticker)

	c.Send("🔬 Running multi\\-agent deep analysis for *"+escapeMD(ticker)+"*\\.\\.\\.",
		&tele.SendOptions{ParseMode: tele.ModeMarkdownV2})
	c.Notify(tele.Typing)

	ar, err := b.client.Analyze(ticker)
	if err != nil {
		return c.Send("❌ "+escapeMD(err.Error()),
			&tele.SendOptions{ParseMode: tele.ModeMarkdownV2, ReplyMarkup: mainKeyboard()})
	}

	msg := b.formatAnalyzeResult(ar)
	return b.sendLongMessage(c, msg, mainKeyboard())
}

func (b *Bot) runSignal(c tele.Context, ticker string) error {
	b.setMode(c.Sender().ID, modeNone)
	ticker = cleanTicker(ticker)

	c.Send("⚡ Getting signal for *"+escapeMD(ticker)+"*\\.\\.\\.",
		&tele.SendOptions{ParseMode: tele.ModeMarkdownV2})
	c.Notify(tele.Typing)

	ar, err := b.client.Analyze(ticker)
	if err != nil {
		return c.Send("❌ "+escapeMD(err.Error()),
			&tele.SendOptions{ParseMode: tele.ModeMarkdownV2, ReplyMarkup: mainKeyboard()})
	}

	msg := b.formatSignalCard(ar)
	return c.Send(msg, &tele.SendOptions{
		ParseMode:   tele.ModeMarkdownV2,
		ReplyMarkup: mainKeyboard(),
	})
}

func (b *Bot) runNews(c tele.Context, ticker string) error {
	b.setMode(c.Sender().ID, modeNone)
	ticker = cleanTicker(ticker)

	c.Send("📰 Fetching news for *"+escapeMD(ticker)+"*\\.\\.\\.",
		&tele.SendOptions{ParseMode: tele.ModeMarkdownV2})
	c.Notify(tele.Typing)

	nr, err := b.client.News(ticker)
	if err != nil {
		return c.Send("❌ "+escapeMD(err.Error()),
			&tele.SendOptions{ParseMode: tele.ModeMarkdownV2, ReplyMarkup: mainKeyboard()})
	}

	if len(nr.Articles) == 0 {
		return c.Send("📭 No recent news for *"+escapeMD(ticker)+"*", &tele.SendOptions{
			ParseMode:   tele.ModeMarkdownV2,
			ReplyMarkup: mainKeyboard(),
		})
	}

	var lines []string
	lines = append(lines, "📰 *"+escapeMD(strings.ToUpper(ticker))+" News*")
	for i, a := range nr.Articles {
		if i >= 5 {
			break
		}
		lines = append(lines, "• "+escapeMD(a.Summary))
	}
	return c.Send(strings.Join(lines, "\n"), &tele.SendOptions{
		ParseMode:   tele.ModeMarkdownV2,
		ReplyMarkup: mainKeyboard(),
	})
}

func (b *Bot) runScan(c tele.Context, market string) error {
	b.setMode(c.Sender().ID, modeNone)

	marketName := "US"
	if market == "hk" {
		marketName = "HK"
	}
	c.Send("📊 Scanning "+marketName+" market\\.\\.\\.", &tele.SendOptions{ParseMode: tele.ModeMarkdownV2})
	c.Notify(tele.Typing)

	resp, err := b.client.Scan(market, "")
	if err != nil {
		return c.Send("❌ "+escapeMD(err.Error()),
			&tele.SendOptions{ParseMode: tele.ModeMarkdownV2, ReplyMarkup: mainKeyboard()})
	}

	return b.sendLongMessage(c, "📊 *"+marketName+" Market Scan*\n\n"+resp.Reply, mainKeyboard())
}

func (b *Bot) runChat(c tele.Context, message string) error {
	sessionID := fmt.Sprintf("tg:%d", c.Sender().ID)

	c.Notify(tele.Typing)

	b.store.SaveMessage(sessionID, "user", message)

	resp, err := b.client.Chat(message, sessionID, false)
	if err != nil {
		return c.Send("❌ "+escapeMD(err.Error()),
			&tele.SendOptions{ParseMode: tele.ModeMarkdownV2, ReplyMarkup: mainKeyboard()})
	}

	b.store.SaveMessage(sessionID, "assistant", resp.Reply)

	display := resp.Reply
	if resp.Signal != nil && resp.Signal.ConfidenceScore > 0 {
		display += b.formatSignalInline(resp.Signal)
	}

	return b.sendLongMessage(c, display, mainKeyboard())
}

// ── Formatting helpers ──

func (b *Bot) formatAnalyzeResult(ar *client.AnalyzeResponse) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📈 *%s*", escapeMD(ar.Ticker)))
	if ar.Price > 0 {
		sb.WriteString(fmt.Sprintf("  \\$%.2f", ar.Price))
	}
	sb.WriteString("\n\n")
	sb.WriteString(ar.Summary)

	if ar.Signal != nil && ar.Signal.ConfidenceScore > 0 {
		sb.WriteString("\n\n")
		sb.WriteString(b.formatSignalInline(ar.Signal))
	}
	return sb.String()
}

func (b *Bot) formatSignalCard(ar *client.AnalyzeResponse) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("⚡ *%s Signal*\n\n", escapeMD(ar.Ticker)))

	if ar.Price > 0 {
		sb.WriteString(fmt.Sprintf("Price: $%.2f\n", ar.Price))
	}

	if ar.Signal != nil && ar.Signal.ConfidenceScore > 0 {
		sb.WriteString(b.formatSignalInline(ar.Signal))
	} else {
		sb.WriteString("No signal available\\. Run *Analyze* for a full report\\.")
	}
	return sb.String()
}

func (b *Bot) formatSignalInline(s *client.SignalData) string {
	emoji := "⚪"
	switch s.DirectionalBias {
	case "Bullish-leaning":
		emoji = "🟢"
	case "Bearish-leaning":
		emoji = "🔴"
	case "Balanced":
		emoji = "🟡"
	}

	strengthBar := ""
	switch s.SignalStrength {
	case "Strong":
		strengthBar = "████"
	case "Moderate":
		strengthBar = "███"
	case "Weak":
		strengthBar = "██"
	default:
		strengthBar = "██"
	}

	return fmt.Sprintf(
		"\\-\\-\\-\n📊 *Signal Card*\n"+
			"%s *%s*  \\|  Confidence: *%d/100*  \\|  %s\n"+
			"⚠️ Not trading advice\\. Directional signal only\\.",
		emoji,
		escapeMD(s.DirectionalBias),
		s.ConfidenceScore,
		strengthBar,
	)
}

// ── Utilities ──

func (b *Bot) sendLongMessage(c tele.Context, text string, kb *tele.ReplyMarkup) error {
	opts := &tele.SendOptions{
		ParseMode:   tele.ModeMarkdownV2,
		ReplyMarkup: kb,
	}
	const maxLen = 4000
	for i := 0; i < len(text); i += maxLen {
		end := i + maxLen
		if end > len(text) {
			end = len(text)
		}
		var err error
		if i == 0 {
			err = c.Send(text[i:end], opts)
		} else {
			// Subsequent chunks: no keyboard to avoid dupes
			_, err = b.tg.Send(c.Recipient(), text[i:end], &tele.SendOptions{
				ParseMode: tele.ModeMarkdownV2,
			})
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func cleanTicker(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "$")
	s = strings.ToUpper(s)
	// Handle "0700.HK" → pass as-is; cloud handles it
	return s
}

func escapeMD(s string) string {
	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	)
	return replacer.Replace(s)
}
