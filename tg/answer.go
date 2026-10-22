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

type AnswerStore interface {
	Accountser
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
	TextChooseBet  = "Выберите ставку:"
	TextAllIn      = "ВА-БАНК!"
)

func (h *AnswerHandler) Handle(ctx context.Context, upd Update) (msgs []Request, err error) {
	qID, answer, err := qIDAndData(OptCallbackDataPrefix, upd.CallbackQuery.Data)
	if err != nil {
		return nil, errors.Join(ErrWrongCallbackData, err)
	}

	accs, err := h.Store.Accounts(ctx, []int{upd.CallbackQuery.Message.Chat.ID})
	if err != nil {
		return nil, fmt.Errorf("accounts: %w", err)
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
				{
					Type:    InputRichBlockButtonsType,
					Align:   AlignLeft,
					Buttons: BetButtons(accs[upd.CallbackQuery.Message.Chat.ID], []int{100, 200, 500}, qID),
				},
			},
		},
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

const TextYourBet = "Ваша ставка: "

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
	if bet == "-100" {
		bet = TextAllIn
	} else {
		bet += "$"
	}
	return InputRichBlock{
		Type: "paragraph",
		Text: TextYourBet + bet,
	}
}

func BetButtons(acc int, bets []int, questionID string) (btns []RichMessageButton) {
	for _, b := range bets {
		if b > acc {
			continue
		}

		btns = append(btns, RichMessageButton{
			Text:         fmt.Sprintf("%d$", b),
			CallbackData: fmt.Sprintf("%s%s:%d", BetCallbackDataPrefix, questionID, b),
		})
	}

	btns = append(btns, RichMessageButton{
		Text: TextAllIn,
		// NOTE: -100 is a special for all-in
		CallbackData: fmt.Sprintf("%s%s:%d", BetCallbackDataPrefix, questionID, -100),
	})

	return btns
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
