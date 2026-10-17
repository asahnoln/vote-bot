package tg

import (
	"context"

	"github.com/asahnoln/vote-bot/store"
)

type RichestUserser interface {
	RichestUsers(ctx context.Context) ([]store.User, error)
}

type StatsRichestHandler struct {
	Store RichestUserser
}

func (h *StatsRichestHandler) Handle(ctx context.Context, upd Update) (msgs []URLer, err error) {
	want := SendRichMessage{
		ChatID: 500,
		RichMessage: InputRichMessage{
			Blocks: []InputRichBlock{
				{
					Type:      "table",
					Caption:   "Самые богатые",
					IsStriped: true,
					Cells: [][]RichBlockTableCell{
						{
							{
								Text:  "Cool name",
								Align: "center",
							},
							{
								Text:  "9999$",
								Align: "center",
							},
						},
						{
							{
								Text:  "Other name",
								Align: "center",
							},
							{
								Text:  "9999$",
								Align: "center",
							},
						},
					},
				},
			},
		},
	}
	msgs = append(msgs, want)
	return msgs, nil
}
