package tg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
)

const (
	OptCallbackDataPrefix Prefix = "opt:"
	BetCallbackDataPrefix Prefix = "bet:"
)

var (
	ErrBetIsNotANumber   = errors.New("bet is not a number")
	ErrWrongCallbackData = errors.New("wrong callback data")
)

type AnswerSaver interface {
	SaveAnswer(ctx context.Context, userID int, questionID string, answer string) error
}

type AnswerHandler struct {
	Store AnswerSaver
}

type AnswerCallbackData struct {
	Type       string
	QuestionID string
	Answer     string
	Question   string
}

func (h *AnswerHandler) Condition(upd Update) bool {
	if upd.CallbackQuery == nil {
		return false
	}

	var d AnswerCallbackData
	err := json.Unmarshal([]byte(upd.CallbackQuery.Data), &d)
	if err != nil {
		slog.Error("answer condition: callback data unmarshal", "err", err)
		return false
	}

	return d.Type == "opt"
}

func (h *AnswerHandler) Name() string {
	return "answer"
}

func (h *AnswerHandler) Handle(ctx context.Context, upd Update) (msgs []URLer, err error) {
	var d AnswerCallbackData
	err = json.Unmarshal([]byte(upd.CallbackQuery.Data), &d)
	if err != nil {
		return nil, fmt.Errorf("callback data unmarshal: %w", err)
	}

	prefix := string(BetCallbackDataPrefix) + d.QuestionID + ":"
	msg := EditMessageText{
		ChatID:    upd.CallbackQuery.Message.Chat.ID,
		MessageID: upd.CallbackQuery.Message.MessageID,
		RichMessage: InputRichMessage{
			Blocks: []InputRichBlock{
				{
					Type: InputRichBlockParagraphType,
					Text: d.Question,
				},
				{
					Type: InputRichBlockParagraphType,
					Text: "Ваш ответ: " + d.Answer,
				},
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

	err = h.Store.SaveAnswer(ctx,
		upd.CallbackQuery.Message.Chat.ID,
		d.QuestionID,
		d.Answer,
	)
	if err != nil {
		return nil, fmt.Errorf("save answer: %w", err)
	}

	msgs = append(msgs, msg)
	return msgs, nil
}

func (h *AnswerHandler) HandleResponse(ctx context.Context, resp Message) error {
	return nil
}

type BetSaver interface {
	SaveBet(ctx context.Context, userID int, questionID string, bet int) error
}

type BetHandler struct {
	Store BetSaver
}

func (h *BetHandler) Condition(upd Update) bool {
	return upd.CallbackQuery != nil && strings.HasPrefix(upd.CallbackQuery.Data, string(BetCallbackDataPrefix))
}

func (h *BetHandler) Name() string {
	return "bet"
}

func (h *BetHandler) Handle(ctx context.Context, upd Update) (msgs []URLer, err error) {
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

	err = h.Store.SaveBet(ctx,
		upd.CallbackQuery.Message.Chat.ID,
		qID,
		bet,
	)
	if err != nil {
		return nil, fmt.Errorf("save bet: %w", err)
	}

	msgs = append(msgs, msg)
	return msgs, nil
}

func (h *BetHandler) HandleResponse(ctx context.Context, resp Message) error {
	return nil
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
