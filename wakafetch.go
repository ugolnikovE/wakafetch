package main

import (
	"fmt"
	"math"
	"os"
	"strings"
	"unicode/utf8"
)

const (
	// labelWidth is the width of the label column in the full output.
	labelWidth = 9
	// barSegments is the number of segments in a language bar.
	barSegments = 10
	// newLineWithIndent starts a new line of the full output.
	newLineWithIndent = "\n  "
	// placeholder is printed in place of a missing value.
	placeholder = "-"
)

// formatDuration formats seconds as hours and minutes, such as "2h 5m".
// The remaining seconds are dropped.
func formatDuration(seconds float64) string {
	var h, m int64
	m = int64(seconds) / 60
	h = m / 60
	m = m - h*60
	return fmt.Sprintf("%dh %dm", h, m)
}

// formatBar formats a language as its name, padded to maxLangLen runes,
// a bar of barSegments segments filled by its percentage, and the
// percentage itself. Percentages outside 0-100 are clamped in the bar.
func formatBar(lang StatUnit, maxLangLen int) string {
	segCount := int(math.Floor(lang.Percent * barSegments / 100.0))

	if segCount < 0 {
		segCount = 0
	} else if segCount > barSegments {
		segCount = barSegments
	}

	var sb strings.Builder

	fmt.Fprintf(&sb, "%-*s  ", maxLangLen, lang.Name)
	fmt.Fprint(&sb, strings.Repeat("█", segCount))
	fmt.Fprint(&sb, strings.Repeat("░", barSegments-segCount))
	fmt.Fprintf(&sb, "  %5.2f%%", lang.Percent)

	return sb.String()
}

// formatLanguageTop returns one row per language, with the bars aligned.
// Only the first row is labeled. If langs is empty, it returns a single
// row with the placeholder.
func formatLanguageTop(langs []StatUnit) []string {
	var langNameLen int
	for _, lang := range langs {
		n := utf8.RuneCountInString(lang.Name)
		if n > langNameLen {
			langNameLen = n
		}
	}

	var langTop []string

	if len(langs) == 0 {
		langTop = append(langTop, formatRow("lang", ""))
		return langTop
	}

	for i, lang := range langs {
		label := ""
		if i == 0 {
			label = "lang"
		}
		langTop = append(langTop, formatRow(label, formatBar(lang, langNameLen)))
	}

	return langTop
}

// formatHeader returns the title line of the full output.
func formatHeader(rng string) string {
	return fmt.Sprintf("wakafetch | %s", rng)
}

// formatFloatStat returns a row with value formatted as a percentage.
func formatFloatStat(label string, value float64) string {
	return fmt.Sprintf("%-*s %.2f%%", labelWidth, label, value)
}

// formatRow returns label padded to labelWidth, followed by value.
// An empty value is replaced with the placeholder.
func formatRow(label string, value string) string {
	if value == "" {
		value = placeholder
	}
	return fmt.Sprintf("%-*s %s", labelWidth, label, value)
}

// formatFullStat returns the multi-line output: a header followed by one
// row per statistic. The number of languages shown is set by cfg.LangTop.
func formatFullStat(cfg Config, stats WakafetchStats) string {
	var sb strings.Builder
	var top int

	if cfg.LangTop.IsAll {
		top = len(stats.Languages)
	} else if cfg.LangTop.Value > len(stats.Languages) {
		top = len(stats.Languages)
	} else {
		top = cfg.LangTop.Value
	}

	fmt.Fprint(&sb, newLineWithIndent+formatHeader(cfg.Range.String()))
	fmt.Fprint(&sb, "\n--------------------------------")
	fmt.Fprint(&sb, newLineWithIndent+formatRow("time", formatDuration(stats.TotalSeconds)))
	fmt.Fprint(&sb, newLineWithIndent+strings.Join(formatLanguageTop(stats.Languages[:top]), newLineWithIndent))
	fmt.Fprint(&sb, newLineWithIndent+formatRow("os", stats.TopOS))
	fmt.Fprint(&sb, newLineWithIndent+formatRow("editor", stats.TopEditor))
	fmt.Fprint(&sb, newLineWithIndent+formatRow("project", stats.TopProject))
	fmt.Fprint(&sb, newLineWithIndent+formatFloatStat("ai", stats.AIPercentage))
	fmt.Fprint(&sb, "\n\n")

	return sb.String()
}

// formatShortStat returns the single-line output: total time,
// top language and AI percentage.
func formatShortStat(stats WakafetchStats) string {
	var lang string
	if len(stats.Languages) > 0 {
		lang = stats.Languages[0].Name
	} else {
		lang = placeholder
	}
	return fmt.Sprintf("%s | %s | ai %.2f%%\n", formatDuration(stats.TotalSeconds), lang, stats.AIPercentage)
}

// Render prints stats to stdout in the format selected by cfg.IsShort.
// It returns an error if writing to stdout fails.
func Render(cfg Config, stats WakafetchStats) error {
	var err error
	if cfg.IsShort {
		_, err = fmt.Fprint(os.Stdout, formatShortStat(stats))
	} else {
		_, err = fmt.Fprint(os.Stdout, formatFullStat(cfg, stats))
	}

	if err != nil {
		return fmt.Errorf("write to stdout: %w", err)
	}

	return nil
}
