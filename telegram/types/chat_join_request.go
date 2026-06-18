package types

import (
	"time"

	"github.com/mtgo-labs/mtgo/tg"
)

// ChatJoinRequest represents a user's request to join a chat via invite link,
// with methods to approve or decline the request.
//
// Example:
//
//	req := types.ParseChatJoinRequest(rawUpdate, users, chats)
//	if err := req.Approve(); err != nil {
//	    log.Fatal(err)
//	}
type ChatJoinRequest struct {
	Chat       *Chat
	FromUser   *User
	Date       time.Time
	Bio        string
	InviteLink *ChatInviteLink
	binder     Binder
}

func (r *ChatJoinRequest) SetBinder(b Binder) {
	r.binder = b
}

func (r *ChatJoinRequest) Approve() error {
	if r.binder == nil {
		return ErrNoBinder
	}
	return r.binder.BoundApproveJoinRequest(r.Chat.ID, r.FromUser.ID)
}

func (r *ChatJoinRequest) Decline() error {
	if r.binder == nil {
		return ErrNoBinder
	}
	return r.binder.BoundDeclineJoinRequest(r.Chat.ID, r.FromUser.ID)
}

// ParseChatJoinRequest converts a TL UpdateBotChatInviteRequester into a ChatJoinRequest.
// Returns nil if raw is nil.
//
// Example:
//
//	req := types.ParseChatJoinRequest(rawUpdate, userMap, chatMap)
//	fmt.Printf("Join request from %s to %s\n", req.FromUser.FirstName, req.Chat.Title)
func ParseChatJoinRequest(
	raw *tg.UpdateBotChatInviteRequester,
	users map[int64]*User,
	chats map[int64]*Chat,
) *ChatJoinRequest {
	if raw == nil {
		return nil
	}
	r := &ChatJoinRequest{
		Date: time.Unix(int64(raw.Date), 0),
		Bio:  raw.About,
	}
	if user := users[raw.UserID]; user != nil {
		r.FromUser = user
	} else {
		r.FromUser = &User{ID: raw.UserID}
	}
	switch p := raw.Peer.(type) {
	case *tg.PeerChat:
		if chat := chats[-p.ChatID]; chat != nil {
			r.Chat = chat
		} else {
			r.Chat = &Chat{ID: -p.ChatID, Type: ChatTypeGroup}
		}
	case *tg.PeerChannel:
		chatID := channelChatID(p.ChannelID)
		if chat := chats[chatID]; chat != nil {
			r.Chat = chat
		} else {
			r.Chat = &Chat{ID: chatID, Type: ChatTypeSupergroup}
		}
	}
	if inv, ok := raw.Invite.(*tg.ChatInviteExported); ok {
		r.InviteLink = ParseChatInviteLink(inv, nil)
	}
	return r
}
