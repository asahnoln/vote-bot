package tg

import (
	"fmt"
	"net/http"
)

const BetCallbackDataPrefix = "bet:"

func (b *Bot) answer(w http.ResponseWriter, r *http.Request, upd Update) (data any, err error) {
	msg := EditMessageText{
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
							CallbackData: BetCallbackDataPrefix + "100",
						},
						{
							Text:         "200",
							CallbackData: BetCallbackDataPrefix + "200",
						},
						{
							Text:         "500",
							CallbackData: BetCallbackDataPrefix + "500",
						},
					},
				},
			},
		},
	}

	err = b.Store.SaveAnswer(r.Context(), upd.CallbackQuery.Message.Chat.ID, upd.CallbackQuery.Data)
	if err != nil {
		return nil, fmt.Errorf("save answer: %w", err)
	}

	return msg, nil
}
