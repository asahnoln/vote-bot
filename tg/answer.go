package tg

import (
	"fmt"
	"net/http"
	"strings"
)

const BetCallbackDataPrefix = "bet:"

func (b *Bot) answer(w http.ResponseWriter, r *http.Request, upd Update) (data URLer, err error) {
	if strings.HasPrefix(upd.CallbackQuery.Data, BetCallbackDataPrefix) {
		msg := EditMessageText{
			ChatID:    upd.CallbackQuery.Message.Chat.ID,
			MessageID: upd.CallbackQuery.Message.MessageID,
			RichMessage: InputRichMessage{
				Blocks: []InputRichBlock{
					{
						Type: "paragraph",
						// TODO: Подумать над ответом
						Text: "Ваш ответ и ваша ставка!",
					},
				},
			},
		}

		err = b.Store.SaveBet(r.Context(),
			upd.CallbackQuery.Message.Chat.ID,
			"bet-q-1",
			500,
		)
		if err != nil {
			return nil, fmt.Errorf("save bet: %w", err)
		}

		return msg, nil
	}

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

	err = b.Store.SaveAnswer(r.Context(),
		upd.CallbackQuery.Message.Chat.ID,
		strings.TrimPrefix(upd.CallbackQuery.Data, OptCallbackDataPrefix))
	if err != nil {
		return nil, fmt.Errorf("save answer: %w", err)
	}

	return msg, nil
}
