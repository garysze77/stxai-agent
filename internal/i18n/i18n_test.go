package i18n

import (
	"testing"
)

func TestT_English(t *testing.T) {
	text := T("display.bull_case", "en")
	if text != "Bull Case" {
		t.Errorf("expected 'Bull Case', got '%s'", text)
	}
}

func TestT_TraditionalChinese(t *testing.T) {
	text := T("display.bull_case", "zh-HK")
	if text != "看好觀點" {
		t.Errorf("expected '看好觀點', got '%s'", text)
	}
}

func TestT_SimplifiedChinese(t *testing.T) {
	text := T("display.bull_case", "zh-CN")
	if text != "看多观点" {
		t.Errorf("expected '看多观点', got '%s'", text)
	}
}

func TestT_FallbackToEnglish(t *testing.T) {
	text := T("display.bull_case", "fr")
	if text != "Bull Case" {
		t.Errorf("expected English fallback 'Bull Case', got '%s'", text)
	}
}

func TestT_MissingKey(t *testing.T) {
	text := T("nonexistent.key", "en")
	if text != "nonexistent.key" {
		t.Errorf("expected key name as fallback, got '%s'", text)
	}
}

func TestT_WithArgs(t *testing.T) {
	text := T("cmd.current_version", "en", "1.0.0")
	if text != "Current version: 1.0.0" {
		t.Errorf("expected 'Current version: 1.0.0', got '%s'", text)
	}
}

func TestT_ChineseWithArgs(t *testing.T) {
	text := T("bot.running_analysis", "zh-HK", "AAPL")
	expected := "🔬 正在為 *AAPL* 執行多智能體深度分析\\.\\.\\。"
	if text != expected {
		t.Errorf("expected '%s', got '%s'", expected, text)
	}
}

func TestSupported(t *testing.T) {
	if !Supported("en") {
		t.Error("en should be supported")
	}
	if !Supported("zh-HK") {
		t.Error("zh-HK should be supported")
	}
	if !Supported("zh-CN") {
		t.Error("zh-CN should be supported")
	}
	if Supported("fr") {
		t.Error("fr should not be supported")
	}
}
