package types

import (
	"testing"

	"github.com/mtgo-labs/mtgo/tg"
)

func TestParseChatJoinRequestUsesNormalizedChatID(t *testing.T) {
	channelID := int64(42)
	chatID := channelChatID(channelID)
	wantChat := &Chat{ID: chatID, Type: ChatTypeSupergroup}
	wantUser := &User{ID: 7, FirstName: "Ada"}

	req := ParseChatJoinRequest(
		&tg.UpdateBotChatInviteRequester{
			Peer:   &tg.PeerChannel{ChannelID: channelID},
			UserID: wantUser.ID,
		},
		map[int64]*User{wantUser.ID: wantUser},
		map[int64]*Chat{chatID: wantChat},
	)

	if req.Chat != wantChat {
		t.Fatalf("Chat = %#v, want normalized map entry %#v", req.Chat, wantChat)
	}
	if req.FromUser != wantUser {
		t.Fatalf("FromUser = %#v, want map entry %#v", req.FromUser, wantUser)
	}
}

func TestParseChatJoinRequestFallsBackToIDs(t *testing.T) {
	req := ParseChatJoinRequest(
		&tg.UpdateBotChatInviteRequester{
			Peer:   &tg.PeerChat{ChatID: 42},
			UserID: 7,
		},
		nil,
		nil,
	)

	if req.Chat == nil || req.Chat.ID != -42 || req.Chat.Type != ChatTypeGroup {
		t.Fatalf("Chat = %#v, want minimal group chat", req.Chat)
	}
	if req.FromUser == nil || req.FromUser.ID != 7 {
		t.Fatalf("FromUser = %#v, want minimal user ID 7", req.FromUser)
	}
}
