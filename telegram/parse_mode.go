package telegram

import "github.com/mtgo-labs/mtgo/telegram/params"

// ParseMode is the text formatting mode for messages.
// Use the short constants: Markdown, HTML, RichMarkdown, RichHTML, MarkdownV2, Disabled.
type ParseMode = params.ParseMode

const (
	// Markdown formats message text as Markdown.
	Markdown = params.Markdown
	// HTML formats message text as HTML.
	HTML = params.HTML
	// RichMarkdown formats text using Telegram's Rich Message Markdown.
	RichMarkdown = params.RichMarkdown
	// RichHTML formats text using Telegram's Rich Message HTML.
	RichHTML = params.RichHTML
	// MarkdownV2 formats message text as MarkdownV2.
	MarkdownV2 = params.MarkdownV2
	// Disabled sends raw text with no formatting.
	Disabled = params.Disabled
)
