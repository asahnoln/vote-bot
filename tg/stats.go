package tg

import (
	"context"
	"fmt"

	"github.com/asahnoln/vote-bot/store"
)

type RichestUserser interface {
	RichestUsers(ctx context.Context) ([]store.User, error)
}

type StatsRichestHandler struct {
	Store RichestUserser
}

func (h *StatsRichestHandler) Handle(ctx context.Context, upd Update) (msgs []URLer, err error) {
	users, err := h.Store.RichestUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("richest users: %w", err)
	}

	msg := SendRichMessage{
		ChatID: 500,
		RichMessage: InputRichMessage{
			Blocks: []InputRichBlock{
				{
					Type:      "table",
					Caption:   "Самые богатые",
					IsStriped: true,
				},
			},
		},
	}

	for _, u := range users {
		msg.RichMessage.Blocks[0].Cells = append(msg.RichMessage.Blocks[0].Cells, []RichBlockTableCell{
			{Text: u.Name, Align: AlignCenter},
			{Text: fmt.Sprintf("%d$", u.Acc), Align: AlignCenter},
		})
	}
	msgs = append(msgs, msg)
	return msgs, nil
}
