package tg

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	OptCallbackDataPrefix Prefix = "opt:"
	BetCallbackDataPrefix Prefix = "bet:"
)

var (
	ErrBetIsNotANumber            = errors.New("bet is not a number")
	ErrWrongCallbackData          = errors.New("wrong callback data")
	ErrNotEnoughOriginalMsgBlocks = errors.New("not enough original message blocks")
)

type AnswerSaver interface {
	SaveAnswer(ctx context.Context, userID int, questionID string, answer string) error
}

type AnswerHandler struct {
	Store AnswerSaver
}

func (h *AnswerHandler) Condition(upd Update) bool {
	return upd.CallbackQuery != nil && strings.HasPrefix(upd.CallbackQuery.Data, string(OptCallbackDataPrefix))
}

func (h *AnswerHandler) Name() string {
	return "answer"
}

const (
	TextYourAnswer = "Ваш ответ: "
	TextChooseBet  = "Выберите ставку:"
)

func (h *AnswerHandler) Handle(ctx context.Context, upd Update) (msgs []URLer, err error) {
	qID, answer, err := qIDAndData(OptCallbackDataPrefix, upd.CallbackQuery.Data)
	if err != nil {
		return nil, errors.Join(ErrWrongCallbackData, err)
	}

	err = h.Store.SaveAnswer(ctx,
		upd.CallbackQuery.Message.Chat.ID,
		qID,
		answer,
	)
	if err != nil {
		return nil, fmt.Errorf("save answer: %w", err)
	}

	if len(upd.CallbackQuery.Message.RichMessage.Blocks) < 2 {
		return nil, ErrNotEnoughOriginalMsgBlocks
	}

	prefix := string(BetCallbackDataPrefix) + qID + ":"
	msg := EditMessageText{
		ChatID:    upd.CallbackQuery.Message.Chat.ID,
		MessageID: upd.CallbackQuery.Message.MessageID,
		RichMessage: InputRichMessage{
			Blocks: []InputRichBlock{
				upd.CallbackQuery.Message.RichMessage.Blocks[0],
				{
					Type: InputRichBlockParagraphType,
					Text: TextYourAnswer + answer,
				},
				{
					Type: InputRichBlockParagraphType,
					Text: TextChooseBet,
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
						{
							Text: "Ва-банк!",
							// NOTE: -100 is a special for all-in
							CallbackData: prefix + "-100",
						},
						// TODO: Add all in
					},
				},
			},
		},
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

const TextYourBet = "Ваша ставка: "

func (h *BetHandler) Handle(ctx context.Context, upd Update) (msgs []URLer, err error) {
	qID, betStr, err := qIDAndData(BetCallbackDataPrefix, upd.CallbackQuery.Data)
	if err != nil {
		return nil, err
	}

	bet, err := strconv.Atoi(betStr)
	if err != nil {
		return nil, errors.Join(ErrBetIsNotANumber, err)
	}

	if len(upd.CallbackQuery.Message.RichMessage.Blocks) < 2 {
		return nil, ErrNotEnoughOriginalMsgBlocks
	}

	msg := EditMessageText{
		ChatID:    upd.CallbackQuery.Message.Chat.ID,
		MessageID: upd.CallbackQuery.Message.MessageID,
		RichMessage: InputRichMessage{
			Blocks: []InputRichBlock{
				upd.CallbackQuery.Message.RichMessage.Blocks[0],
				upd.CallbackQuery.Message.RichMessage.Blocks[1],
				BetInputRichBlock(betStr),
			},
		},
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

func BetInputRichBlock(bet string) InputRichBlock {
	if bet == "-100" {
		bet = "ВА-БАНК!"
	}
	return InputRichBlock{
		Type: "paragraph",
		Text: TextYourBet + bet,
	}
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
