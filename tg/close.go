package tg

import (
	"context"
	"fmt"

	"github.com/asahnoln/vote-bot/store"
)

type LostUserSaver interface {
	SaveLostUser(ctx context.Context, userID int) error
}

type Moder interface {
	Mode(context.Context) (store.Mode, error)
}

type CloseStore interface {
	Userser
	CurrentQuestioner
	Answerser
	LostUserSaver
	Moder
	CloseCurrentQuestion(ctx context.Context) error
}

type CloseHandler struct {
	Store CloseStore
}

func (h *CloseHandler) HandleResponse(context.Context, any) error {
	return nil
}

func (h *CloseHandler) Name() string {
	return "close"
}

func (h *CloseHandler) Condition(upd Update) bool {
	return upd.Message.Text == CloseCurrentQuestionCommand
}

func (h *CloseHandler) Handle(ctx context.Context, upd Update) (msgs []Request, err error) {
	users, err := h.Store.Users(ctx)
	if err != nil {
		return nil, fmt.Errorf("users: %w", err)
	}

	q, err := h.Store.CurrentQuestion(ctx)
	if err != nil {
		return nil, fmt.Errorf("current question: %w", err)
	}

	answers, err := h.Store.Answers(ctx, q.ID)
	if err != nil {
		return nil, fmt.Errorf("answers: %w", err)
	}

	answersMap := make(map[int]store.Answer)
	for _, a := range answers {
		answersMap[a.UserID] = a
	}

	err = h.Store.CloseCurrentQuestion(ctx)
	if err != nil {
		return nil, fmt.Errorf("close current question: %w", err)
	}

	mode, err := h.Store.Mode(ctx)
	if err != nil {
		return nil, fmt.Errorf("mode: %w", err)
	}

	for _, u := range users {
		a := answersMap[u.UserID]
		msgs = append(msgs,
			ClosingMessage(
				u.UserID,
				u.LastMessageID,
				q.Q,
				a,
				mode,
			),
		)

		if mode != store.TutorialMode && a.Bet == 0 { // Works for both missing answer and bet
			err := h.Store.SaveLostUser(ctx, u.UserID)
			if err != nil {
				return nil, fmt.Errorf("save lost user %d: %w", u.UserID, err)
			}
		}
	}
	return msgs, nil
}

const (
	TextQuestionClosed          = "Вопрос закрыт. Скоро узнаем, насколько вы были правы."
	TextYouMissedBet            = "Увы, вы не успели сделать ставку. Ваша игра окончена. "
	TextByeBye                  = "Казино прощается с вами!"
	TextYouMissedAnswer         = "Вопрос закрыт! Вы не успели ответить. По правилам казино вы покидаете игру."
	TextYouMissedBetTutorial    = "Увы, вы не успели сделать ставку. Ваша игра была бы окончена, но это всего лишь обучение. "
	TextByeByeTutorial          = "Казино еще не прощается с вами! Ждите следующего вопроса."
	TextYouMissedAnswerTutorial = "Вопрос закрыт! Вы не успели ответить. По правилам казино вы бы покинули игру, но это всего лишь обучение."
)

func ClosingMessage(userID int, messageID int, question string, a store.Answer, mode store.Mode) (msg EditMessageText) {
	msg = EditMessageText{
		ChatID:    userID,
		MessageID: messageID,
		RichMessage: InputRichMessage{
			Blocks: []InputRichBlock{
				{
					Type: InputRichBlockParagraphType,
					Text: question,
				},
			},
		},
	}

	if a.A != "" {
		msg.RichMessage.Blocks = append(msg.RichMessage.Blocks, InputRichBlock{
			Type: InputRichBlockParagraphType,
			Text: TextYourAnswer + a.A,
		})

		if a.Bet > 0 {
			betBlock := InputRichBlock{
				Type: InputRichBlockParagraphType,
			}
			if a.AllIn {
				betBlock.Text = TextYourBet + TextAllIn
			} else {
				betBlock.Text = fmt.Sprintf("%s%d$", TextYourBet, a.Bet)
			}

			msg.RichMessage.Blocks = append(msg.RichMessage.Blocks, betBlock)
			msg.RichMessage.Blocks = append(msg.RichMessage.Blocks, InputRichBlock{
				Type: InputRichBlockParagraphType,
				Text: TextQuestionClosed,
			})
		} else {
			betBlock := InputRichBlock{
				Type: InputRichBlockParagraphType,
				Text: TextYouMissedBet,
			}
			byeBlock := InputRichBlock{
				Type: InputRichBlockParagraphType,
				Text: TextByeBye,
			}
			if mode == store.TutorialMode {
				betBlock.Text = TextYouMissedBetTutorial
				byeBlock.Text = TextByeByeTutorial
			}

			msg.RichMessage.Blocks = append(msg.RichMessage.Blocks, betBlock)
			msg.RichMessage.Blocks = append(msg.RichMessage.Blocks, byeBlock)
		}
	} else {
		answerBlock := InputRichBlock{
			Type: InputRichBlockParagraphType,
			Text: TextYouMissedAnswer,
		}
		if mode == store.TutorialMode {
			answerBlock.Text = TextYouMissedAnswerTutorial
		}
		msg.RichMessage.Blocks = append(msg.RichMessage.Blocks, answerBlock)
	}

	return
}
