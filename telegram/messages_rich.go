package telegram

import (
	"context"
	"fmt"

	"github.com/mtgo-labs/mtgo/telegram/params"
	"github.com/mtgo-labs/mtgo/telegram/types"
	"github.com/mtgo-labs/mtgo/tg"
)

// SendRichMessageDraft streams a temporary rich message preview to a private chat.
// The preview is ephemeral and disappears after 30 seconds unless it is updated.
// Reuse the same non-zero draftID for consecutive updates so Telegram can animate
// transitions between them. Call [Client.SendRichMessage] to persist the final
// rich message; sending a draft alone does not create a message in chat history.
//
// Parameters:
//   - ctx: context for cancellation and deadlines
//   - chatID: the target private chat
//   - draftID: a non-zero identifier shared by all updates of the same draft
//   - text: the partial rich document to display
//   - parseMode: optional Rich Markdown or Rich HTML mode; defaults to Rich Markdown
//
// Returns an error if the draft ID is zero, the text is empty, the mode is
// unsupported, the peer
// cannot be resolved, or the RPC call fails.
func (c *Client) SendRichMessageDraft(ctx context.Context, chatID, draftID int64, text string, parseMode ...params.ParseMode) error {
	mode := params.ParseModeDefault
	if len(parseMode) > 1 {
		return fmt.Errorf("too many parse modes: expected 0 or 1, got %d", len(parseMode))
	}
	if len(parseMode) == 1 {
		mode = parseMode[0]
	}
	richMessage, err := richMessageFromText(text, mode)
	if err != nil {
		return err
	}
	return c.SendInputRichMessageDraft(ctx, chatID, draftID, richMessage)
}

// SendInputRichMessageDraft sends a prebuilt low-level Rich Message draft.
func (c *Client) SendInputRichMessageDraft(ctx context.Context, chatID, draftID int64, richMessage tg.InputRichMessageClass) error {
	c.Log.Debugf("SendRichMessageDraft chat_id=%d draft_id=%d", chatID, draftID)
	if draftID == 0 {
		return fmt.Errorf("rich message draft ID cannot be zero")
	}
	if richMessage == nil {
		return fmt.Errorf("rich message cannot be nil")
	}

	return c.SendChatAction(ctx, chatID, &tg.InputSendMessageRichMessageDraftAction{
		RandomID:    draftID,
		RichMessage: richMessage,
	})
}

// SendRichMessage sends text as a Telegram Rich Message. ParseMode defaults to
// Rich Markdown and supports ParseModeRichMarkdown and ParseModeRichHTML.
func (c *Client) SendRichMessage(ctx context.Context, chatID int64, text string, opts ...*params.SendMessage) (*types.Message, error) {
	opt := params.GetOptDef(&params.SendMessage{}, opts...)
	richMessage, err := richMessageFromText(text, opt.ParseMode)
	if err != nil {
		return nil, err
	}
	return c.SendInputRichMessage(ctx, chatID, richMessage, opt)
}

// SendInputRichMessage sends a prebuilt low-level Rich Message.
func (c *Client) SendInputRichMessage(ctx context.Context, chatID int64, richMessage tg.InputRichMessageClass, opts ...*params.SendMessage) (*types.Message, error) {
	c.Log.Debugf("SendRichMessage chat_id=%d", chatID)
	if richMessage == nil {
		return nil, fmt.Errorf("rich message cannot be nil")
	}

	peer, err := resolvePeer(c, chatID)
	if err != nil {
		peer, err = c.ResolvePeer(ctx, chatID)
		if err != nil && len(opts) > 0 {
			opt := opts[0]
			if opt.ReplyToMessageID != 0 && chatID > 0 {
				peer = &tg.InputPeerUserFromMessage{
					Peer:   &tg.InputPeerUser{UserID: chatID},
					MsgID:  opt.ReplyToMessageID,
					UserID: chatID,
				}
				err = nil
			}
		}
		if err != nil {
			return nil, fmt.Errorf("resolve peer: %w", err)
		}
	}
	opt := params.GetOptDef(&params.SendMessage{}, opts...)

	var flags tg.Fields
	if opt.DisableWebPagePreview {
		flags.Set(1)
	}
	if opt.Silent || opt.DisableNotification {
		flags.Set(5)
	}
	if opt.Background {
		flags.Set(6)
	}
	if opt.ClearDraft {
		flags.Set(7)
	}
	if opt.NoForwards {
		flags.Set(14)
	}
	if opt.InvertMedia {
		flags.Set(16)
	}

	var replyTo tg.InputReplyToClass
	if opt.ReplyTo != nil {
		flags.Set(0)
		replyTo = opt.ReplyTo
	} else if opt.ReplyToMessageID != 0 {
		flags.Set(0)
		replyTo = &tg.InputReplyToMessage{ReplyToMsgID: opt.ReplyToMessageID}
	}
	if opt.ReplyMarkup != nil {
		flags.Set(2)
	}
	flags.Set(23)
	if opt.SendAs != nil {
		flags.Set(13)
	}

	req := &tg.MessagesSendMessageRequest{
		Flags:       flags,
		Silent:      opt.Silent || opt.DisableNotification,
		Background:  opt.Background,
		ClearDraft:  opt.ClearDraft,
		Noforwards:  opt.NoForwards,
		InvertMedia: opt.InvertMedia,
		Peer:        peer,
		ReplyTo:     replyTo,
		Message:     "",
		RandomID:    c.RandomID(),
		ReplyMarkup: opt.ReplyMarkup,
		Entities:    nil,
		SendAs:      opt.SendAs,
		RichMessage: richMessage,
	}
	if opt.ScheduleDate != nil {
		req.ScheduleDate = *opt.ScheduleDate
	}
	if opt.EffectID != nil {
		req.Effect = *opt.EffectID
	}

	result, err := c.Raw().MessagesSendMessage(ctx, req)
	if err != nil {
		return nil, err
	}
	return extractSingleMessage(result, c)
}

// EditRichMessage replaces an existing message with Rich Markdown text.
func (c *Client) EditRichMessage(ctx context.Context, chatID int64, messageID int32, text string, opts ...*params.EditMessage) (*types.Message, error) {
	opt := params.GetOptDef(&params.EditMessage{}, opts...)
	richMessage, err := richMessageFromText(text, opt.ParseMode)
	if err != nil {
		return nil, err
	}
	return c.EditInputRichMessage(ctx, chatID, messageID, richMessage, opt)
}

// EditInputRichMessage edits a message using a prebuilt low-level Rich Message.
func (c *Client) EditInputRichMessage(ctx context.Context, chatID int64, messageID int32, richMessage tg.InputRichMessageClass, opts ...*params.EditMessage) (*types.Message, error) {
	c.Log.Debugf("EditRichMessage chat_id=%d msg_id=%d", chatID, messageID)
	if richMessage == nil {
		return nil, fmt.Errorf("rich message cannot be nil")
	}

	peer, err := resolvePeer(c, chatID)
	if err != nil {
		return nil, fmt.Errorf("resolve peer: %w", err)
	}
	opt := params.GetOptDef(&params.EditMessage{}, opts...)

	var flags tg.Fields
	if opt.DisableWebPagePreview {
		flags.Set(1)
	}
	if opt.ReplyMarkup != nil {
		flags.Set(2)
	}
	if opt.InvertMedia {
		flags.Set(16)
	}
	flags.Set(23)

	req := &tg.MessagesEditMessageRequest{
		Flags:       flags,
		InvertMedia: opt.InvertMedia,
		Peer:        peer,
		ID:          messageID,
		Message:     "",
		ReplyMarkup: opt.ReplyMarkup,
		Entities:    nil,
		RichMessage: richMessage,
	}
	if opt.ScheduleDate != nil {
		req.ScheduleDate = *opt.ScheduleDate
	}

	result, err := c.Raw().MessagesEditMessage(ctx, req)
	if err != nil {
		return nil, err
	}
	return extractSingleMessage(result, c)
}

func richMessageFromText(text string, mode params.ParseMode) (tg.InputRichMessageClass, error) {
	if text == "" {
		return nil, fmt.Errorf("rich message text cannot be empty")
	}
	switch mode {
	case "", params.ParseModeDefault, params.ParseModeRichMarkdown:
		return &tg.InputRichMessageMarkdown{Markdown: text}, nil
	case params.ParseModeRichHTML:
		return &tg.InputRichMessageHTML{HTML: text}, nil
	default:
		return nil, fmt.Errorf("unsupported rich message parse mode %q", mode)
	}
}
