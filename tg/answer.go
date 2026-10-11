package tg

import (
	"fmt"
	"net/http"
)

func (b *Bot) answer(w http.ResponseWriter, r *http.Request, upd Update) (msg EditMessageText, err error) {
	msg = EditMessageText{
		ChatID:    upd.CallbackQuery.Message.Chat.ID,
		MessageID: upd.CallbackQuery.Message.MessageID,
		RichMessage: InputRichMessage{
			Blocks: []InputRichBlock{
				{
					Type: "paragraph",
					Text: "Выберите ставку:",
				},
				{
					Type:  "buttons",
					Align: "left",
					Buttons: []RichMessageButton{
						{
							Text:         "100",
							CallbackData: "100",
						},
						{
							Text:         "200",
							CallbackData: "200",
						},
						{
							Text:         "500",
							CallbackData: "500",
						},
					},
				},
			},
		},
	}

	err = b.Store.SaveAnswer(r.Context(), upd.CallbackQuery.Message.Chat.ID, upd.CallbackQuery.Data)
	if err != nil {
		return msg, fmt.Errorf("save answer: %w", err)
	}

	return msg, nil
}
