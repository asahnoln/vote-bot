package tg

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

const BetCallbackDataPrefix = "bet:"

func (b *Bot) answer(w http.ResponseWriter, r *http.Request, upd Update) (data URLer, err error) {
	// TODO: Test for wrong data format
	qID, answer, _ := qIDAndData(OptCallbackDataPrefix, upd.CallbackQuery.Data)

	prefix := BetCallbackDataPrefix + qID + ":"
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
							CallbackData: prefix + "100",
						},
						{
							Text:         "200",
							CallbackData: prefix + "200",
						},
						{
							Text:         "500",
							CallbackData: prefix + "500",
						},
					},
				},
			},
		},
	}

	err = b.Store.SaveAnswer(r.Context(),
		upd.CallbackQuery.Message.Chat.ID,
		qID,
		answer,
	)
	if err != nil {
		return nil, fmt.Errorf("save answer: %w", err)
	}

	return msg, nil
}

var (
	ErrBetIsNotANumber   = errors.New("bet is not a number")
	ErrWrongCallbackData = errors.New("wrong callback data")
)

func (b *Bot) bet(w http.ResponseWriter, r *http.Request, upd Update) (data URLer, err error) {
	msg := EditMessageText{
		ChatID:    upd.CallbackQuery.Message.Chat.ID,
		MessageID: upd.CallbackQuery.Message.MessageID,
		RichMessage: InputRichMessage{
			Blocks: []InputRichBlock{
				{
					Type: InputRichBlockParagraphType,
					// TODO: Подумать над ответом
					Text: "Ваш ответ и ваша ставка приняты!",
				},
			},
		},
	}

	qID, betStr, err := qIDAndData(BetCallbackDataPrefix, upd.CallbackQuery.Data)
	if err != nil {
		return nil, err
	}

	bet, err := strconv.Atoi(betStr)
	if err != nil {
		return nil, errors.Join(ErrBetIsNotANumber, err)
	}

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

func qIDAndData(prefix Prefix, s string) (qID string, data string, err error) {
	qIDandDataStr := strings.TrimPrefix(s, string(prefix))

	colPos := strings.IndexRune(qIDandDataStr, ':')
	if colPos == -1 {
		return "", "", ErrWrongCallbackData
	}

	qID = qIDandDataStr[:colPos]
	data = qIDandDataStr[colPos+1:]

	return
}
