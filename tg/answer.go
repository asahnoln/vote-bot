package tg

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/asahnoln/vote-bot/store"
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

type AnswerStore interface {
	User(ctx context.Context, userID int) (store.User, error)
	SaveAnswer(ctx context.Context, userID int, questionID string, answer string) error
}

type AnswerHandler struct {
	Store AnswerStore
}

func (h *AnswerHandler) Condition(upd Update) bool {
	return upd.CallbackQuery != nil && strings.HasPrefix(upd.CallbackQuery.Data, string(OptCallbackDataPrefix))
}

func (h *AnswerHandler) Name() string {
	return "answer"
}

const (
	TextYourAnswer = "Ваш ответ: "
	TextChooseBet  = "Ответ принят! Теперь решайте, сколько поставить на кон."
	TextAllIn      = "ВА-БАНК!"
)

func (h *AnswerHandler) Handle(ctx context.Context, upd Update) (msgs []Request, err error) {
	qID, answer, err := qIDAndData(OptCallbackDataPrefix, upd.CallbackQuery.Data)
	if err != nil {
		return nil, errors.Join(ErrWrongCallbackData, err)
	}

	user, err := h.Store.User(ctx, upd.CallbackQuery.Message.Chat.ID)
	if err != nil {
		return nil, fmt.Errorf("user: %w", err)
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
			},
		},
	}

	btnsRows := BetButtonsRows(user.Acc, []int{100, 200, 500}, qID)
	for _, r := range btnsRows {
		msg.RichMessage.Blocks = append(msg.RichMessage.Blocks, InputRichBlock{
			Type:    InputRichBlockButtonsType,
			Align:   AlignLeft,
			Buttons: r,
		})
	}

	msgs = append(msgs, msg)
	return msgs, nil
}

func (h *AnswerHandler) HandleResponse(ctx context.Context, resp any) error {
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

const (
	TextYourBet     = "Ваша ставка: "
	TextBetAccepted = "Ставка принята!"
)

func (h *BetHandler) Handle(ctx context.Context, upd Update) (msgs []Request, err error) {
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
				{
					Type: InputRichBlockParagraphType,
					Text: TextBetAccepted,
				},
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

func (h *BetHandler) HandleResponse(ctx context.Context, resp any) error {
	return nil
}

func BetInputRichBlock(bet string) InputRichBlock {
	return InputRichBlock{
		Type: "paragraph",
		Text: TextYourBet + bet + "$",
	}
}

func BetButtonsRows(acc int, bets []int, questionID string) (btnsRows [][]RichMessageButton) {
	btnsRows = make([][]RichMessageButton, 1)
	for _, b := range bets {
		if b >= acc {
			continue
		}

		btnsRows[0] = append(btnsRows[0], RichMessageButton{
			Text:         fmt.Sprintf("%d$", b),
			CallbackData: fmt.Sprintf("%s%s:%d", BetCallbackDataPrefix, questionID, b),
		})
	}

	if len(btnsRows[0]) == 0 {
		btnsRows = make([][]RichMessageButton, 0)
	}

	btnsRows = append(btnsRows, []RichMessageButton{
		{
			Text:         fmt.Sprintf("%s (%d$)", TextAllIn, acc),
			CallbackData: fmt.Sprintf("%s%s:%d", BetCallbackDataPrefix, questionID, acc),
		},
	})

	return btnsRows
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
