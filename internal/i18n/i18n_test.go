package i18n

import (
	"testing"
)

func TestT_English(t *testing.T) {
	text := T("cli.errors.missing_key", "en")
	if text == "cli.errors.missing_key" {
		t.Error("expected translation, got key back")
	}
	if text == "" {
		t.Error("expected non-empty translation")
	}
}

func TestT_ChineseHK(t *testing.T) {
	text := T("cli.errors.missing_key", "zh-HK")
	if text == "cli.errors.missing_key" {
		t.Error("expected translation, got key back")
	}
	if text == "" {
		t.Error("expected non-empty translation")
	}
}

func TestT_ChineseCN(t *testing.T) {
	text := T("cli.errors.missing_key", "zh-CN")
	if text == "cli.errors.missing_key" {
		t.Error("expected translation, got key back")
	}
	if text == "" {
		t.Error("expected non-empty translation")
	}
}

func TestT_FallbackToEnglish(t *testing.T) {
	// Invalid lang should fall back to English
	text := T("cli.errors.missing_key", "fr")
	enText := T("cli.errors.missing_key", "en")
	if text != enText {
		t.Errorf("expected fallback to English, got %q vs %q", text, enText)
	}
}

func TestT_MissingKeyReturnsKey(t *testing.T) {
	text := T("nonexistent.key.xyz", "en")
	if text != "nonexistent.key.xyz" {
		t.Errorf("expected key name for missing key, got %s", text)
	}
}

func TestT_ParameterSubstitution(t *testing.T) {
	text := T("cli.errors.command_error", "en", "test error message")
	if text == "cli.errors.command_error" {
		t.Error("expected parameter substitution")
	}
	if text == T("cli.errors.command_error", "en") {
		t.Error("expected different text with and without args")
	}
}

func TestValidLang(t *testing.T) {
	if !ValidLang("en") {
		t.Error("en should be valid")
	}
	if !ValidLang("zh-HK") {
		t.Error("zh-HK should be valid")
	}
	if !ValidLang("zh-CN") {
		t.Error("zh-CN should be valid")
	}
	if ValidLang("fr") {
		t.Error("fr should not be valid")
	}
}

func TestT_AllDisplayKeys(t *testing.T) {
	keys := []string{
		"cli.display.bull_case",
		"cli.display.bear_case",
		"cli.display.synthesis",
		"cli.display.signal",
		"cli.display.direction",
		"cli.display.confidence",
		"cli.display.strength",
		"cli.display.generating",
		"cli.display.error_in",
	}
	for _, lang := range []string{"en", "zh-HK", "zh-CN"} {
		for _, key := range keys {
			text := T(key, lang)
			if text == key {
				t.Errorf("missing translation: %s in %s", key, lang)
			}
		}
	}
}
