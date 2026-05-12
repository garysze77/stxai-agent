package bot

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/garysze77/stxai-agent/internal/client"
	"github.com/garysze77/stxai-agent/internal/i18n"
	"github.com/garysze77/stxai-agent/internal/store"

	tele "gopkg.in/telebot.v3"
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
	lang   string

	mu           sync.Mutex
	modes        map[int64]string
	allowedUsers map[int64]bool
}

func New(token string, c *client.Client, s *store.Store, lang string, allowedUsers []int64) (*Bot, error) {
	if lang == "" {
		lang = "en"
	}
	pref := tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	tg, err := tele.NewBot(pref)
	if err != nil {
		return nil, fmt.Errorf("%s", i18n.T("bot.bot_init_failed", lang, err))
	}

	au := make(map[int64]bool, len(allowedUsers))
	for _, id := range allowedUsers {
		au[id] = true
	}

	return &Bot{
		tg:           tg,
		client:       c,
		store:        s,
		lang:         lang,
		modes:        make(map[int64]string),
		allowedUsers: au,
	}, nil
}

// Button helpers — keyboard button labels stay in bot default lang
// so handleText comparisons work regardless of user language.
func (b *Bot) btnAnalyze() string { return i18n.T("bot.btn_analyze", b.lang) }
func (b *Bot) btnSignal() string  { return i18n.T("bot.btn_signal", b.lang) }
func (b *Bot) btnNews() string    { return i18n.T("bot.btn_news", b.lang) }
func (b *Bot) btnScanUS() string  { return i18n.T("bot.btn_scan_us", b.lang) }
func (b *Bot) btnScanHK() string  { return i18n.T("bot.btn_scan_hk", b.lang) }
func (b *Bot) btnClear() string   { return i18n.T("bot.btn_clear", b.lang) }
func (b *Bot) btnHelp() string    { return i18n.T("bot.btn_help", b.lang) }
func (b *Bot) btnLang() string   { return i18n.T("bot.btn_lang", b.lang) }
func (b *Bot) btnPair() string   { return i18n.T("bot.btn_pair", b.lang) }
func (b *Bot) btnCancel() string { return i18n.T("bot.btn_cancel", b.lang) }

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

// ── Authorization ──

func (b *Bot) isAllowed(userID int64) bool {
	if len(b.allowedUsers) == 0 {
		return true
	}
	return b.allowedUsers[userID]
}

func (b *Bot) getUserLang(userID int64) string {
	lang, err := b.store.GetPreference(fmt.Sprintf("lang:%d", userID))
	if err != nil || lang == "" {
		return b.lang
	}
	return lang
}

func (b *Bot) langFor(c tele.Context) string {
	return b.getUserLang(c.Sender().ID)
}

func (b *Bot) denyAccess(c tele.Context) error {
	return c.Send(i18n.T("bot.unauthorized", b.langFor(c)), &tele.SendOptions{
		ParseMode: tele.ModeMarkdownV2,
	})
}

// ── Keyboard ──

func (b *Bot) mainKeyboard() *tele.ReplyMarkup {
	m := &tele.ReplyMarkup{ResizeKeyboard: true}
	m.ReplyKeyboard = [][]tele.ReplyButton{
		{{Text: b.btnAnalyze()}, {Text: b.btnSignal()}, {Text: b.btnNews()}},
		{{Text: b.btnScanUS()}, {Text: b.btnScanHK()}},
		{{Text: b.btnLang()}, {Text: b.btnPair()}},
		{{Text: b.btnClear()}, {Text: b.btnHelp()}},
	}
	return m
}

func (b *Bot) cancelKeyboard() *tele.ReplyMarkup {
	m := &tele.ReplyMarkup{ResizeKeyboard: true}
	m.ReplyKeyboard = [][]tele.ReplyButton{
		{{Text: b.btnCancel()}},
	}
	return m
}

// ── Startup ──

func (b *Bot) Start(ctx context.Context) error {
	b.registerHandlers()
	b.setCommands()
	log.Printf("%s", i18n.T("bot.starting", b.lang, b.tg.Me.Username))
	b.tg.Start()
	<-ctx.Done()
	b.tg.Stop()
	return nil
}

func (b *Bot) setCommands() {
	cmds := []tele.Command{
		{Text: "analyze", Description: i18n.T("bot.cmd_analyze_desc", b.lang)},
		{Text: "scan", Description: i18n.T("bot.cmd_scan_desc", b.lang)},
		{Text: "news", Description: i18n.T("bot.cmd_news_desc", b.lang)},
		{Text: "signal", Description: i18n.T("bot.cmd_signal_desc", b.lang)},
		{Text: "clear", Description: i18n.T("bot.cmd_clear_desc", b.lang)},
		{Text: "help", Description: i18n.T("bot.cmd_help_desc", b.lang)},
		{Text: "lang", Description: i18n.T("bot.cmd_lang_desc", b.lang)},
		{Text: "pair", Description: i18n.T("bot.cmd_pair_desc", b.lang)},
	}
	if err := b.tg.SetCommands(cmds); err != nil {
		log.Printf("%s", i18n.T("bot.set_commands_failed", b.lang, err))
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
	b.tg.Handle("/lang", b.handleLang)
	b.tg.Handle("/pair", b.handlePair)
	b.tg.Handle(tele.OnText, b.handleText)
	b.tg.Handle(tele.OnCallback, b.handleCallback)
}

// ── Slash command handlers ──

func (b *Bot) handleStart(c tele.Context) error {
	if !b.isAllowed(c.Sender().ID) {
		return b.denyAccess(c)
	}
	b.setMode(c.Sender().ID, modeNone)
	return c.Send(i18n.T("bot.welcome", b.langFor(c)), &tele.SendOptions{
		ParseMode:   tele.ModeMarkdownV2,
		ReplyMarkup: b.mainKeyboard(),
	})
}

func (b *Bot) handleHelp(c tele.Context) error {
	if !b.isAllowed(c.Sender().ID) {
		return b.denyAccess(c)
	}
	b.setMode(c.Sender().ID, modeNone)
	return c.Send(i18n.T("bot.help_msg", b.langFor(c)), &tele.SendOptions{
		ParseMode:   tele.ModeMarkdownV2,
		ReplyMarkup: b.mainKeyboard(),
	})
}

func (b *Bot) handleAnalyze(c tele.Context) error {
	if !b.isAllowed(c.Sender().ID) {
		return b.denyAccess(c)
	}
	ticker := strings.TrimSpace(c.Message().Payload)
	if ticker == "" {
		return b.promptTicker(c, modeAnalyze, i18n.T("bot.prompt_analyze", b.langFor(c)))
	}
	return b.runAnalyze(c, ticker)
}

func (b *Bot) handleSignal(c tele.Context) error {
	if !b.isAllowed(c.Sender().ID) {
		return b.denyAccess(c)
	}
	ticker := strings.TrimSpace(c.Message().Payload)
	if ticker == "" {
		return b.promptTicker(c, modeSignal, i18n.T("bot.prompt_signal", b.langFor(c)))
	}
	return b.runSignal(c, ticker)
}

func (b *Bot) handleNews(c tele.Context) error {
	if !b.isAllowed(c.Sender().ID) {
		return b.denyAccess(c)
	}
	ticker := strings.TrimSpace(c.Message().Payload)
	if ticker == "" {
		return b.promptTicker(c, modeNews, i18n.T("bot.prompt_news", b.langFor(c)))
	}
	return b.runNews(c, ticker)
}

func (b *Bot) handleScan(c tele.Context) error {
	if !b.isAllowed(c.Sender().ID) {
		return b.denyAccess(c)
	}
	market := strings.TrimSpace(c.Message().Payload)
	if market == "" {
		market = "us"
	}
	if market != "us" && market != "hk" {
		return c.Send(i18n.T("bot.scan_usage", b.langFor(c)),
			&tele.SendOptions{ParseMode: tele.ModeMarkdownV2, ReplyMarkup: b.mainKeyboard()})
	}
	return b.runScan(c, market)
}

func (b *Bot) handleClear(c tele.Context) error {
	if !b.isAllowed(c.Sender().ID) {
		return b.denyAccess(c)
	}
	b.setMode(c.Sender().ID, modeNone)
	b.store.DeleteOldMessages(time.Now())
	return c.Send(i18n.T("bot.history_cleared", b.langFor(c)), &tele.SendOptions{
		ParseMode:   tele.ModeMarkdownV2,
		ReplyMarkup: b.mainKeyboard(),
	})
}

func (b *Bot) handleLang(c tele.Context) error {
	if !b.isAllowed(c.Sender().ID) {
		return b.denyAccess(c)
	}
	userLang := b.getUserLang(c.Sender().ID)

	inlineKeys := [][]tele.InlineButton{
		{
			{Text: "English", Data: "lang:en"},
			{Text: "繁體中文", Data: "lang:zh-HK"},
			{Text: "简体中文", Data: "lang:zh-CN"},
		},
	}
	return c.Send(i18n.T("bot.select_lang", userLang), &tele.SendOptions{
		ParseMode: tele.ModeMarkdownV2,
		ReplyMarkup: &tele.ReplyMarkup{
			InlineKeyboard: inlineKeys,
		},
	})
}

func (b *Bot) handlePair(c tele.Context) error {
	if !b.isAllowed(c.Sender().ID) {
		return b.denyAccess(c)
	}
	userLang := b.getUserLang(c.Sender().ID)

	err := b.client.PairLink(c.Sender().ID, c.Sender().Username)
	if err != nil {
		log.Printf("PairLink failed: %v", err)
		return c.Send("❌ "+escapeMD(err.Error()), &tele.SendOptions{
			ParseMode:   tele.ModeMarkdownV2,
			ReplyMarkup: b.mainKeyboard(),
		})
	}

	return c.Send(i18n.T("bot.pair_success", userLang), &tele.SendOptions{
		ParseMode:   tele.ModeMarkdownV2,
		ReplyMarkup: b.mainKeyboard(),
	})
}

func (b *Bot) handleCallback(c tele.Context) error {
	data := c.Data()
	if strings.HasPrefix(data, "lang:") {
		lang := strings.TrimPrefix(data, "lang:")
		userID := c.Sender().ID
		b.store.SetPreference(fmt.Sprintf("lang:%d", userID), lang)
		b.setMode(userID, modeNone)
		return c.Send(i18n.T("bot.lang_set", lang), &tele.SendOptions{
			ParseMode:   tele.ModeMarkdownV2,
			ReplyMarkup: b.mainKeyboard(),
		})
	}
	return c.Respond()
}

// ── Keyboard button / text handler ──

func (b *Bot) handleText(c tele.Context) error {
	if !b.isAllowed(c.Sender().ID) {
		return b.denyAccess(c)
	}
	text := c.Text()
	userID := c.Sender().ID
	userLang := b.getUserLang(userID)

	// 1. Route slash commands that may arrive as plain text
	if strings.HasPrefix(text, "/") {
		switch {
		case text == "/pair" || strings.HasPrefix(text, "/pair "):
			return b.handlePair(c)
		case text == "/lang" || strings.HasPrefix(text, "/lang "):
			return b.handleLang(c)
		case text == "/start" || strings.HasPrefix(text, "/start "):
			return b.handleStart(c)
		case text == "/help" || strings.HasPrefix(text, "/help "):
			return b.handleHelp(c)
		case text == "/clear" || strings.HasPrefix(text, "/clear "):
			return b.handleClear(c)
		case text == "/analyze" || strings.HasPrefix(text, "/analyze "):
			return b.handleAnalyze(c)
		case text == "/scan" || strings.HasPrefix(text, "/scan "):
			return b.handleScan(c)
		case text == "/news" || strings.HasPrefix(text, "/news "):
			return b.handleNews(c)
		case text == "/signal" || strings.HasPrefix(text, "/signal "):
			return b.handleSignal(c)
		}
	}

	// 2. Check if it's a keyboard button press
	switch text {
	case b.btnAnalyze():
		return b.promptTicker(c, modeAnalyze, i18n.T("bot.prompt_analyze", userLang))
	case b.btnSignal():
		return b.promptTicker(c, modeSignal, i18n.T("bot.prompt_signal", userLang))
	case b.btnNews():
		return b.promptTicker(c, modeNews, i18n.T("bot.prompt_news", userLang))
	case b.btnScanUS():
		return b.runScan(c, "us")
	case b.btnScanHK():
		return b.runScan(c, "hk")
	case b.btnLang():
		return b.handleLang(c)
	case b.btnPair():
		return b.handlePair(c)
	case b.btnClear():
		return b.handleClear(c)
	case b.btnHelp():
		return b.handleHelp(c)
	case b.btnCancel():
		b.setMode(userID, modeNone)
		return c.Send(i18n.T("bot.cancelled", userLang), &tele.SendOptions{
			ParseMode:   tele.ModeMarkdownV2,
			ReplyMarkup: b.mainKeyboard(),
		})
	}

	// 3. Check if user is in a mode (waiting for ticker input)
	mode := b.getMode(userID)
	switch mode {
	case modeAnalyze:
		return b.runAnalyze(c, text)
	case modeSignal:
		return b.runSignal(c, text)
	case modeNews:
		return b.runNews(c, text)
	}

	// 4. Normal chat — send to cloud API
	return b.runChat(c, text)
}

// ── Prompt for ticker ──

func (b *Bot) promptTicker(c tele.Context, mode, msg string) error {
	b.setMode(c.Sender().ID, mode)
	return c.Send(msg, &tele.SendOptions{
		ParseMode:   tele.ModeMarkdownV2,
		ReplyMarkup: b.cancelKeyboard(),
	})
}

// ── Action runners ──

func (b *Bot) runAnalyze(c tele.Context, ticker string) error {
	b.setMode(c.Sender().ID, modeNone)
	ticker = cleanTicker(ticker)
	lang := b.langFor(c)

	c.Send(i18n.T("bot.running_analysis", lang, escapeMD(ticker)),
		&tele.SendOptions{ParseMode: tele.ModeMarkdownV2})
	c.Notify(tele.Typing)

	ar, err := b.client.Analyze(ticker, lang)
	if err != nil {
		return c.Send("❌ "+escapeMD(err.Error()),
			&tele.SendOptions{ParseMode: tele.ModeMarkdownV2, ReplyMarkup: b.mainKeyboard()})
	}

	msg := b.formatAnalyzeResult(ar, lang)
	return b.sendLongMessage(c, msg, b.mainKeyboard())
}

func (b *Bot) runSignal(c tele.Context, ticker string) error {
	b.setMode(c.Sender().ID, modeNone)
	ticker = cleanTicker(ticker)
	lang := b.langFor(c)

	c.Send(i18n.T("bot.getting_signal", lang, escapeMD(ticker)),
		&tele.SendOptions{ParseMode: tele.ModeMarkdownV2})
	c.Notify(tele.Typing)

	ar, err := b.client.Analyze(ticker, lang)
	if err != nil {
		return c.Send("❌ "+escapeMD(err.Error()),
			&tele.SendOptions{ParseMode: tele.ModeMarkdownV2, ReplyMarkup: b.mainKeyboard()})
	}

	msg := b.formatSignalCard(ar, lang)
	return c.Send(msg, &tele.SendOptions{
		ParseMode:   tele.ModeMarkdownV2,
		ReplyMarkup: b.mainKeyboard(),
	})
}

func (b *Bot) runNews(c tele.Context, ticker string) error {
	b.setMode(c.Sender().ID, modeNone)
	ticker = cleanTicker(ticker)
	lang := b.langFor(c)

	c.Send(i18n.T("bot.fetching_news", lang, escapeMD(ticker)),
		&tele.SendOptions{ParseMode: tele.ModeMarkdownV2})
	c.Notify(tele.Typing)

	nr, err := b.client.News(ticker, lang)
	if err != nil {
		return c.Send("❌ "+escapeMD(err.Error()),
			&tele.SendOptions{ParseMode: tele.ModeMarkdownV2, ReplyMarkup: b.mainKeyboard()})
	}

	if len(nr.Articles) == 0 {
		return c.Send(i18n.T("bot.no_news_for_ticker", lang, escapeMD(ticker)), &tele.SendOptions{
			ParseMode:   tele.ModeMarkdownV2,
			ReplyMarkup: b.mainKeyboard(),
		})
	}

	var lines []string
	lines = append(lines, i18n.T("bot.news_header", lang, escapeMD(strings.ToUpper(ticker))))
	for i, a := range nr.Articles {
		if i >= 5 {
			break
		}
		lines = append(lines, "• "+escapeMD(a.Summary))
	}
	return c.Send(strings.Join(lines, "\n"), &tele.SendOptions{
		ParseMode:   tele.ModeMarkdownV2,
		ReplyMarkup: b.mainKeyboard(),
	})
}

func (b *Bot) runScan(c tele.Context, market string) error {
	b.setMode(c.Sender().ID, modeNone)
	lang := b.langFor(c)

	marketName := "US"
	if market == "hk" {
		marketName = "HK"
	}
	c.Send(i18n.T("bot.scanning_market", lang, marketName),
		&tele.SendOptions{ParseMode: tele.ModeMarkdownV2})
	c.Notify(tele.Typing)

	resp, err := b.client.Scan(market, "", lang)
	if err != nil {
		return c.Send("❌ "+escapeMD(err.Error()),
			&tele.SendOptions{ParseMode: tele.ModeMarkdownV2, ReplyMarkup: b.mainKeyboard()})
	}

	return b.sendLongMessage(c, i18n.T("bot.market_scan_header", lang, marketName)+"\n\n"+resp.Reply, b.mainKeyboard())
}

func (b *Bot) runChat(c tele.Context, message string) error {
	sessionID := fmt.Sprintf("tg:%d", c.Sender().ID)
	lang := b.langFor(c)

	c.Notify(tele.Typing)

	b.store.SaveMessage(sessionID, "user", message)

	resp, err := b.client.Chat(message, sessionID, lang, false)
	if err != nil {
		return c.Send("❌ "+escapeMD(err.Error()),
			&tele.SendOptions{ParseMode: tele.ModeMarkdownV2, ReplyMarkup: b.mainKeyboard()})
	}

	b.store.SaveMessage(sessionID, "assistant", resp.Reply)

	display := resp.Reply
	if resp.Signal != nil && resp.Signal.ConfidenceScore > 0 {
		display += b.formatSignalInline(resp.Signal, lang)
	}

	return b.sendLongMessage(c, display, b.mainKeyboard())
}

// ── Formatting helpers ──

func (b *Bot) formatAnalyzeResult(ar *client.AnalyzeResponse, lang string) string {
	var sb strings.Builder
	sb.WriteString(i18n.T("bot.ticker_analyze_header", lang, escapeMD(ar.Ticker)))
	if ar.Price > 0 {
		sb.WriteString(fmt.Sprintf("  \\$%.2f", ar.Price))
	}
	sb.WriteString("\n\n")
	sb.WriteString(ar.Summary)

	if ar.Signal != nil && ar.Signal.ConfidenceScore > 0 {
		sb.WriteString("\n\n")
		sb.WriteString(b.formatSignalInline(ar.Signal, lang))
	}
	return sb.String()
}

func (b *Bot) formatSignalCard(ar *client.AnalyzeResponse, lang string) string {
	var sb strings.Builder
	sb.WriteString(i18n.T("bot.signal_header", lang, escapeMD(ar.Ticker)))
	sb.WriteString("\n\n")

	if ar.Price > 0 {
		sb.WriteString(fmt.Sprintf(i18n.T("bot.price", lang)+": $%.2f\n", ar.Price))
	}

	if ar.Signal != nil && ar.Signal.ConfidenceScore > 0 {
		sb.WriteString(b.formatSignalInline(ar.Signal, lang))
	} else {
		sb.WriteString(i18n.T("bot.no_signal_available", lang))
	}
	return sb.String()
}

func (b *Bot) formatSignalInline(s *client.SignalData, lang string) string {
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
		"\\-\\-\\-\n"+i18n.T("bot.signal_card", lang)+"\n"+
			"%s *%s*  \\|  "+i18n.T("display.confidence", lang)+": *%d/100*  \\|  %s\n"+
			i18n.T("bot.not_advice", lang),
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
