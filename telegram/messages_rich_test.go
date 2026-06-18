package telegram

import (
	"context"
	"testing"

	"github.com/mtgo-labs/mtgo/telegram/params"
	"github.com/mtgo-labs/mtgo/tg"
)

func TestSendRichMessageDraft(t *testing.T) {
	c, inv := newClientWithMock(t)
	c.CachePeer(10, &tg.InputPeerUser{UserID: 10, AccessHash: 20})
	markdown := "**Working...**"

	if err := c.SendRichMessageDraft(context.Background(), 10, 12345, markdown); err != nil {
		t.Fatalf("SendRichMessageDraft() error: %v", err)
	}

	req, ok := inv.lastCall().(*tg.MessagesSetTypingRequest)
	if !ok {
		t.Fatalf("expected MessagesSetTypingRequest, got %T", inv.lastCall())
	}
	action, ok := req.Action.(*tg.InputSendMessageRichMessageDraftAction)
	if !ok {
		t.Fatalf("expected InputSendMessageRichMessageDraftAction, got %T", req.Action)
	}
	rich, ok := action.RichMessage.(*tg.InputRichMessageMarkdown)
	if !ok || action.RandomID != 12345 || rich.Markdown != markdown {
		t.Fatalf("unexpected rich message draft action: %#v", action)
	}
}

func TestSendRichHTMLDraft(t *testing.T) {
	c, inv := newClientWithMock(t)
	c.CachePeer(10, &tg.InputPeerUser{UserID: 10, AccessHash: 20})
	html := "<h1>Working...</h1>"

	if err := c.SendRichMessageDraft(context.Background(), 10, 12345, html, params.ParseModeRichHTML); err != nil {
		t.Fatalf("SendRichMessageDraft() error: %v", err)
	}

	req, ok := inv.lastCall().(*tg.MessagesSetTypingRequest)
	if !ok {
		t.Fatalf("expected MessagesSetTypingRequest, got %T", inv.lastCall())
	}
	action, ok := req.Action.(*tg.InputSendMessageRichMessageDraftAction)
	if !ok {
		t.Fatalf("expected InputSendMessageRichMessageDraftAction, got %T", req.Action)
	}
	rich, ok := action.RichMessage.(*tg.InputRichMessageHTML)
	if !ok || rich.HTML != html {
		t.Fatalf("unexpected rich HTML draft action: %#v", action)
	}
}

func TestSendRichMessageDraftRejectsInvalidInput(t *testing.T) {
	c, inv := newClientWithMock(t)
	c.CachePeer(10, &tg.InputPeerUser{UserID: 10, AccessHash: 20})

	if err := c.SendRichMessageDraft(context.Background(), 10, 0, "**Working...**"); err == nil {
		t.Fatal("expected zero draft ID error")
	}
	if err := c.SendRichMessageDraft(context.Background(), 10, 12345, ""); err == nil {
		t.Fatal("expected empty rich message error")
	}
	if err := c.SendRichMessageDraft(context.Background(), 10, 12345, "text", params.ParseModeRichMarkdown, params.ParseModeRichHTML); err == nil {
		t.Fatal("expected too many parse modes error")
	}
	if err := c.SendInputRichMessageDraft(context.Background(), 10, 12345, nil); err == nil {
		t.Fatal("expected nil rich message error")
	}
	if inv.callCount() != 0 {
		t.Fatalf("unexpected RPC calls: %d", inv.callCount())
	}
}

func TestSendRichMessage(t *testing.T) {
	c, inv := newClientWithMock(t)
	c.CachePeer(10, &tg.InputPeerChannel{ChannelID: 10, AccessHash: 20})
	scheduleDate := int32(12345)
	markup := &tg.ReplyInlineMarkup{}
	markdown := "# Report\n\n**Ready**"

	_, err := c.SendRichMessage(
		context.Background(),
		10,
		markdown,
		&params.SendMessage{
			Silent:           true,
			ReplyToMessageID: 42,
			ReplyMarkup:      markup,
			ParseMode:        params.ParseModeRichMarkdown,
			ScheduleDate:     &scheduleDate,
		},
	)
	if err != nil {
		t.Fatalf("SendRichMessage() error: %v", err)
	}

	req, ok := inv.lastCall().(*tg.MessagesSendMessageRequest)
	if !ok {
		t.Fatalf("expected MessagesSendMessageRequest, got %T", inv.lastCall())
	}
	rich, ok := req.RichMessage.(*tg.InputRichMessageMarkdown)
	if !ok || rich.Markdown != markdown || !req.Flags.Has(23) {
		t.Fatalf("unexpected rich message request: %#v", req)
	}
	if req.Message != "" || req.Entities != nil || req.Flags.Has(3) {
		t.Fatalf("rich message unexpectedly contains legacy content: message=%q entities=%#v flags=%v", req.Message, req.Entities, req.Flags)
	}
	if !req.Silent || req.ReplyTo == nil || req.ReplyMarkup != markup || req.ScheduleDate != scheduleDate {
		t.Fatalf("existing send options were not preserved: %#v", req)
	}
}

func TestSendRichHTMLMessage(t *testing.T) {
	c, inv := newClientWithMock(t)
	c.CachePeer(10, &tg.InputPeerChannel{ChannelID: 10, AccessHash: 20})
	html := "<h1>Report</h1>"

	if _, err := c.SendRichMessage(context.Background(), 10, html, &params.SendMessage{
		ParseMode: params.ParseModeRichHTML,
	}); err != nil {
		t.Fatalf("SendRichMessage() error: %v", err)
	}

	req, ok := inv.lastCall().(*tg.MessagesSendMessageRequest)
	if !ok {
		t.Fatalf("expected MessagesSendMessageRequest, got %T", inv.lastCall())
	}
	rich, ok := req.RichMessage.(*tg.InputRichMessageHTML)
	if !ok || rich.HTML != html || !req.Flags.Has(23) {
		t.Fatalf("unexpected rich HTML request: %#v", req)
	}
}

func TestSendRichMessageIgnoresLegacyEntities(t *testing.T) {
	c, inv := newClientWithMock(t)
	c.CachePeer(10, &tg.InputPeerChannel{ChannelID: 10, AccessHash: 20})

	if _, err := c.SendRichMessage(context.Background(), 10, "# Report", &params.SendMessage{
		ParseMode: params.ParseModeRichMarkdown,
		Entities:  []tg.MessageEntityClass{&tg.MessageEntityBold{Offset: 0, Length: 6}},
	}); err != nil {
		t.Fatalf("SendRichMessage() error: %v", err)
	}

	req := inv.lastCall().(*tg.MessagesSendMessageRequest)
	if req.Message != "" || req.Entities != nil || req.Flags.Has(3) {
		t.Fatalf("unexpected legacy fallback: message=%q entities=%#v flags=%v", req.Message, req.Entities, req.Flags)
	}
}

func TestEditRichMessage(t *testing.T) {
	c, inv := newClientWithMock(t)
	c.CachePeer(10, &tg.InputPeerChannel{ChannelID: 10, AccessHash: 20})
	markdown := "# Report"

	_, err := c.EditRichMessage(context.Background(), 10, 55, markdown, &params.EditMessage{
		ParseMode: params.ParseModeRichMarkdown,
		Entities:  []tg.MessageEntityClass{&tg.MessageEntityBold{Offset: 0, Length: 6}},
	})
	if err != nil {
		t.Fatalf("EditRichMessage() error: %v", err)
	}

	req, ok := inv.lastCall().(*tg.MessagesEditMessageRequest)
	if !ok {
		t.Fatalf("expected MessagesEditMessageRequest, got %T", inv.lastCall())
	}
	rich, ok := req.RichMessage.(*tg.InputRichMessageMarkdown)
	if !ok || rich.Markdown != markdown || !req.Flags.Has(23) || req.Flags.Has(11) || req.Flags.Has(3) {
		t.Fatalf("unexpected rich edit request: %#v", req)
	}
	if req.Message != "" || req.Entities != nil {
		t.Fatalf("rich edit unexpectedly contains legacy content: message=%q entities=%#v", req.Message, req.Entities)
	}
}

func TestEditRichHTMLMessage(t *testing.T) {
	c, inv := newClientWithMock(t)
	c.CachePeer(10, &tg.InputPeerChannel{ChannelID: 10, AccessHash: 20})
	html := "<h1>Report</h1>"

	if _, err := c.EditRichMessage(context.Background(), 10, 55, html, &params.EditMessage{
		ParseMode: params.ParseModeRichHTML,
	}); err != nil {
		t.Fatalf("EditRichMessage() error: %v", err)
	}

	req, ok := inv.lastCall().(*tg.MessagesEditMessageRequest)
	if !ok {
		t.Fatalf("expected MessagesEditMessageRequest, got %T", inv.lastCall())
	}
	rich, ok := req.RichMessage.(*tg.InputRichMessageHTML)
	if !ok || rich.HTML != html || !req.Flags.Has(23) {
		t.Fatalf("unexpected rich HTML edit request: %#v", req)
	}
}

func TestRichMessageRejectsUnsupportedParseMode(t *testing.T) {
	c, inv := newClientWithMock(t)
	c.CachePeer(10, &tg.InputPeerChannel{ChannelID: 10, AccessHash: 20})

	if _, err := c.SendRichMessage(context.Background(), 10, "# Report", &params.SendMessage{
		ParseMode: params.ParseModeHTML,
	}); err == nil {
		t.Fatal("expected unsupported rich message parse mode error")
	}
	if _, err := c.EditRichMessage(context.Background(), 10, 55, "# Report", &params.EditMessage{
		ParseMode: params.ParseModeMarkdown,
	}); err == nil {
		t.Fatal("expected unsupported rich message parse mode error")
	}
	if inv.callCount() != 0 {
		t.Fatalf("unexpected RPC calls: %d", inv.callCount())
	}
}

func TestRichMessageRejectsEmptyText(t *testing.T) {
	c, inv := newClientWithMock(t)
	c.CachePeer(10, &tg.InputPeerChannel{ChannelID: 10, AccessHash: 20})

	if _, err := c.SendRichMessage(context.Background(), 10, ""); err == nil {
		t.Fatal("expected empty rich message error")
	}
	if _, err := c.EditRichMessage(context.Background(), 10, 55, ""); err == nil {
		t.Fatal("expected empty rich message error")
	}
	if inv.callCount() != 0 {
		t.Fatalf("unexpected RPC calls: %d", inv.callCount())
	}
}

func TestInputRichMessageRejectsNil(t *testing.T) {
	c, inv := newClientWithMock(t)
	c.CachePeer(10, &tg.InputPeerChannel{ChannelID: 10, AccessHash: 20})

	if _, err := c.SendInputRichMessage(context.Background(), 10, nil); err == nil {
		t.Fatal("expected SendInputRichMessage() error")
	}
	if _, err := c.EditInputRichMessage(context.Background(), 10, 55, nil); err == nil {
		t.Fatal("expected EditInputRichMessage() error")
	}
	if inv.callCount() != 0 {
		t.Fatalf("unexpected RPC calls: %d", inv.callCount())
	}
}
