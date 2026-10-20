package tg

import (
	"context"
	"fmt"

	"github.com/asahnoln/vote-bot/store"
)

type StatsRichestStore interface {
	AllUsers(ctx context.Context) ([]store.User, error)
	RichestUsers(ctx context.Context) ([]store.User, error)
}

type StatsRichestHandler struct {
	Store StatsRichestStore
}

func (h *StatsRichestHandler) Name() string {
	return "stats_richestusers"
}

const TextRichest = "Самые богатые"

func (h *StatsRichestHandler) Condition(upd Update) bool {
	return upd.Message.Text == StatsRichestUsersCommand
}

func (h *StatsRichestHandler) Handle(ctx context.Context, upd Update) (msgs []URLer, err error) {
	allUsers, _ := h.Store.AllUsers(ctx)

	richestUsers, err := h.Store.RichestUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("richest users: %w", err)
	}

	msg := SendRichMessage{
		RichMessage: InputRichMessage{
			Blocks: []InputRichBlock{
				{
					Type:      "table",
					Caption:   TextRichest,
					IsStriped: true,
				},
			},
		},
	}

	for _, u := range richestUsers {
		msg.RichMessage.Blocks[0].Cells = append(msg.RichMessage.Blocks[0].Cells, []RichBlockTableCell{
			{Text: u.Name, Align: AlignCenter},
			{Text: fmt.Sprintf("%d$", u.Acc), Align: AlignCenter},
		})
	}

	for _, u := range allUsers {
		msg.ChatID = u.UserID
		msgs = append(msgs, msg)
	}

	return msgs, nil
}

func (h *StatsRichestHandler) HandleResponse(context.Context, Message) error {
	return nil
}
