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

	_, answer := qIDAndData(OptCallbackDataPrefix, upd.CallbackQuery.Data)

	err = b.Store.SaveAnswer(r.Context(),
		upd.CallbackQuery.Message.Chat.ID,
		answer,
	)
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

	qID, betStr := qIDAndData(BetCallbackDataPrefix, upd.CallbackQuery.Data)

	// TODO: Case for wrong bet string
	bet, _ := strconv.Atoi(betStr)

	err = b.Store.SaveBet(r.Context(),
		upd.CallbackQuery.Message.Chat.ID,
		qID,
		bet,
	)
	if err != nil {
		return nil, fmt.Errorf("save bet: %w", err)
	}

	return msg, nil
}

func qIDAndData(prefix Prefix, s string) (qID string, data string) {
	qIDandDataStr := strings.TrimPrefix(s, string(prefix))

	// TODO: Case for wrong format
	colPos := strings.IndexRune(qIDandDataStr, ':')
	qID = qIDandDataStr[:colPos]
	data = qIDandDataStr[colPos+1:]

	return
}
