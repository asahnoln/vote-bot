package tg

import (
	"context"
	"fmt"

	"github.com/asahnoln/vote-bot/store"
)

type Answerser interface {
	Answers(ctx context.Context, questionID string) ([]store.Answer, error)
}

type Accountser interface {
	Accounts(ctx context.Context, userIDs []int) (map[int]int, error)
}

type RevealStore interface {
	Answerser
	LostUserSaver
	Accountser
	LastClosedQuestion(context.Context) (store.Question, error)
}

type RevealHandler struct {
	Store RevealStore
}

func (h *RevealHandler) HandleResponse(context.Context, any) error {
	return nil
}

func (h *RevealHandler) Name() string {
	return "reveal"
}

func (h *RevealHandler) Condition(upd Update) bool {
	return upd.Message.Text == SendAnswerCommand
}

func (h *RevealHandler) Handle(ctx context.Context, upd Update) (msgs []Request, err error) {
	q, err := h.Store.LastClosedQuestion(ctx)
	if err != nil {
		return nil, fmt.Errorf("last closed question: %w", err)
	}

	answers, err := h.Store.Answers(ctx, q.ID)
	if err != nil {
		return nil, fmt.Errorf("answers: %w", err)
	}

	userIDs := make([]int, 0)
	for _, a := range answers {
		// Filter out from response messages
		if a.Bet == 0 {
			continue
		}

		userIDs = append(userIDs, a.UserID)
	}

	accs, err := h.Store.Accounts(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("account: %w", err)
	}

	for _, a := range answers {
		acc, ok := accs[a.UserID]
		// If bet was not set - user was filtered out above
		if !ok {
			continue
		}

		msgs = append(
			msgs,
			RevealMessage(
				a.UserID,
				a.IsRight,
				acc,
			),
		)

		if acc <= 0 {
			err := h.Store.SaveLostUser(ctx, a.UserID)
			if err != nil {
				return nil, fmt.Errorf("save lost user %d: %w", a.UserID, err)
			}
		}
	}

	return msgs, nil
}

const (
	TextYouAreRight    = "Браво! Правильный ответ. Удача на вашей стороне."
	TextYouAreWrongBut = "Не угадали! Вы теряете свою ставку. Посмотрим, удастся ли отыграться."
	TextYouAreWrong    = "Неверный ответ."
	TextYourBalance    = "Ваш счет — "
	TextYouLost        = "Вы банкрот. Казино прощается с вами!"
)

func RevealMessage(userID int, isRight bool, acc int) (msg SendRichMessage) {
	msg = SendRichMessage{
		ChatID: userID,
	}

	if acc > 0 {
		rightMsg := TextYouAreRight
		if !isRight {
			rightMsg = TextYouAreWrongBut
		}
		msg.RichMessage.Blocks = append(msg.RichMessage.Blocks, InputRichBlock{
			Type: InputRichBlockParagraphType,
			Text: rightMsg,
		})
		msg.RichMessage.Blocks = append(msg.RichMessage.Blocks, InputRichBlock{
			Type: InputRichBlockParagraphType,
			Text: fmt.Sprintf("%s%d$", TextYourBalance, acc),
		})

	} else {
		msg.RichMessage.Blocks = append(msg.RichMessage.Blocks, InputRichBlock{
			Type: InputRichBlockParagraphType,
			Text: TextYouAreWrong,
		})
		msg.RichMessage.Blocks = append(msg.RichMessage.Blocks, InputRichBlock{
			Type: InputRichBlockParagraphType,
			Text: fmt.Sprintf("%s%d$", TextYourBalance, 0),
		})
		msg.RichMessage.Blocks = append(msg.RichMessage.Blocks, InputRichBlock{
			Type: InputRichBlockParagraphType,
			Text: TextYouLost,
		})
	}

	return
}
