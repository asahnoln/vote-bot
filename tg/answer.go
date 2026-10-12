package tg

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

const BetCallbackDataPrefix = "bet:"

func (b *Bot) answer(w http.ResponseWriter, r *http.Request, upd Update) (data URLer, err error) {
	msg := EditMessageText{
		ChatID:    upd.CallbackQuery.Message.Chat.ID,
		MessageID: upd.CallbackQuery.Message.MessageID,
		RichMessage: InputRichMessage{
			Blocks: []InputRichBlock{
				{
					Type: InputRichBlockParagraphType,
					Text: "Выберите ставку:",
				},
				{
					Type:  InputRichBlockButtonsType,
					Align: AlignLeft,
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

func (b *Bot) bet(w http.ResponseWriter, r *http.Request, upd Update) (data URLer, err error) {
	msg := EditMessageText{
		ChatID:    upd.CallbackQuery.Message.Chat.ID,
		MessageID: upd.CallbackQuery.Message.MessageID,
		RichMessage: InputRichMessage{
			Blocks: []InputRichBlock{
				{
					Type: InputRichBlockParagraphType,
					// TODO: Подумать над ответом
					Text: "Ваш ответ и ваша ставка!",
				},
			},
		},
	}

	qIDandBetStr := strings.TrimPrefix(upd.CallbackQuery.Data, BetCallbackDataPrefix)

	// TODO: Case for wrong format
	qIDandBet := strings.Split(qIDandBetStr, ":")

	// TODO: Case for wrong bet string
	bet, _ := strconv.Atoi(qIDandBet[1])

	err = b.Store.SaveBet(r.Context(),
		upd.CallbackQuery.Message.Chat.ID,
		qIDandBet[0],
		bet,
	)
	if err != nil {
		return nil, fmt.Errorf("save bet: %w", err)
	}

	return msg, nil
}
