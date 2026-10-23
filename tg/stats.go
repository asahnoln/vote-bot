package tg

import (
	"context"
	"fmt"

	"github.com/asahnoln/vote-bot/store"
)

type StatsStore interface {
	AllUsers(ctx context.Context) ([]store.User, error)
	Stats(ctx context.Context) ([]store.User, error)
}

type StatsHandler struct {
	Store StatsStore
}

func (h *StatsHandler) Name() string {
	return "stats"
}

const TextWinners = "Победители"

func (h *StatsHandler) Condition(upd Update) bool {
	return upd.Message.Text == StatsRichestUsersCommand
}

func (h *StatsHandler) Handle(ctx context.Context, upd Update) (msgs []Request, err error) {
	allUsers, err := h.Store.AllUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("all users: %w", err)
	}

	stats, err := h.Store.Stats(ctx)
	if err != nil {
		return nil, fmt.Errorf("richest users: %w", err)
	}

	msg := SendRichMessage{
		RichMessage: InputRichMessage{
			Blocks: []InputRichBlock{
				{
					Type:      "table",
					Caption:   TextWinners,
					IsStriped: true,
					Cells: [][]RichBlockTableCell{
						{
							{
								IsHeader: true,
								Text:     "Имя",
							},
							{
								IsHeader: true,
								Text:     "Кол-во правильных ответов",
							},
							{
								IsHeader: true,
								Text:     "Баланс",
							},
							{
								IsHeader: true,
								Text:     "Среднее время ответа (в секундах)",
							},
						},
					},
				},
			},
		},
	}

	for _, u := range stats {
		msg.RichMessage.Blocks[0].Cells = append(msg.RichMessage.Blocks[0].Cells, []RichBlockTableCell{
			{Text: u.Name, Align: AlignCenter},
			{Text: fmt.Sprintf("%d", u.RightAnswerCnt), Align: AlignCenter},
			{Text: fmt.Sprintf("%d$", u.Acc), Align: AlignCenter},
			{Text: fmt.Sprintf("%0.2f", u.AvgSpeed), Align: AlignCenter},
		})
	}

	for _, u := range allUsers {
		msg.ChatID = u.UserID
		msgs = append(msgs, msg)
	}

	return msgs, nil
}

func (h *StatsHandler) HandleResponse(context.Context, any) error {
	return nil
}
