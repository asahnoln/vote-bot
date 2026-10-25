package tg_test

import (
	"context"
	"errors"
	"testing"

	"github.com/asahnoln/vote-bot/store"
	"github.com/asahnoln/vote-bot/tg"
	"github.com/google/go-cmp/cmp"
)

// TODO: Scope to admin
func TestCloseHandler(t *testing.T) {
	spy := &closeStoreSpy{
		usersStoreStub: usersStoreStub{
			users: []store.User{
				{UserID: 91, LastMessageID: 67521},
				{UserID: 56, LastMessageID: 303},
				{UserID: 23, LastMessageID: 10002},
				{UserID: 24, LastMessageID: 10003},
			},
		},
		answersStoreStub: answersStoreStub{
			answers: []store.Answer{
				{UserID: 91, A: "Кто-то", Bet: 900},
				{UserID: 56, A: "Кто-то", Bet: 500},
				{UserID: 23, A: "Кто-то"},
			},
		},
	}
	h := tg.CloseHandler{
		Store: spy,
	}

	msgs, err := h.Handle(context.Background(), tg.Update{})
	if err != nil {
		t.Fatalf("got err %v; want nil", err)
	}

	if got, want := spy.closed, true; got != want {
		t.Errorf("question closed: got %v; want %v", got, want)
	}

	if got, want := spy.answersStoreStub.questionID, "opt-q-99"; got != want {
		t.Errorf("answers question: got %v; want %v", got, want)
	}

	if got, want := len(msgs), 4; got != want {
		t.Errorf("msgs count: got %v; want %v", got, want)
	}

	want := tg.EditMessageText{
		ChatID:    91,
		MessageID: 67521,
		RichMessage: tg.InputRichMessage{
			Blocks: []tg.InputRichBlock{
				{
					Type: "paragraph",
					Text: "X or Y?",
				},
				{
					Type: "paragraph",
					Text: "Ваш ответ: Кто-то",
				},
				{
					Type: "paragraph",
					Text: "Ваша ставка: 900$",
				},
				{
					Type: "paragraph",
					Text: "Вопрос закрыт. Скоро узнаем, насколько вы были правы.",
				},
			},
		},
	}
	if diff := cmp.Diff(want, msgs[0]); diff != "" {
		t.Errorf("request mismatch (-want +got):%s", diff)
	}

	if diff := cmp.Diff([]int{23, 24}, spy.lostUserStoreStub.userIDs); diff != "" {
		t.Errorf("lost user IDs mismatch (-want +got):%s", diff)
	}
}

func TestCloseHandlerTutorialWontSaveLosingUser(t *testing.T) {
	spy := &closeStoreSpy{
		usersStoreStub: usersStoreStub{
			users: []store.User{
				{UserID: 24, LastMessageID: 10003},
			},
		},
		mode: store.TutorialMode,
	}
	h := tg.CloseHandler{
		Store: spy,
	}

	msgs, err := h.Handle(context.Background(), tg.Update{})
	if err != nil {
		t.Fatalf("got err %v; want nil", err)
	}

	if got, want := len(msgs), 1; got != want {
		t.Errorf("msgs count: got %v; want %v", got, want)
	}
	if got, want := len(spy.lostUserStoreStub.userIDs), 0; got != want {
		t.Errorf("lost userIDs count: got %v; want %v", got, want)
	}
}

func TestClosingMessage(t *testing.T) {
	msg := tg.ClosingMessage(191, 167521, "Q or W?", store.Answer{A: "WWW!!!", Bet: 1900}, "")
	want := tg.EditMessageText{
		ChatID:    191,
		MessageID: 167521,
		RichMessage: tg.InputRichMessage{
			Blocks: []tg.InputRichBlock{
				{
					Type: "paragraph",
					Text: "Q or W?",
				},
				{
					Type: "paragraph",
					Text: "Ваш ответ: WWW!!!",
				},
				{
					Type: "paragraph",
					Text: "Ваша ставка: 1900$",
				},
				{
					Type: "paragraph",
					Text: "Вопрос закрыт. Скоро узнаем, насколько вы были правы.",
				},
			},
		},
	}
	if diff := cmp.Diff(want, msg); diff != "" {
		t.Errorf("request mismatch (-want +got):%s", diff)
	}
}

func TestClosingMessageAllIn(t *testing.T) {
	msg := tg.ClosingMessage(191, 167521, "Q or W?", store.Answer{A: "WWW!!!", Bet: 999, AllIn: true}, "")
	want := tg.EditMessageText{
		ChatID:    191,
		MessageID: 167521,
		RichMessage: tg.InputRichMessage{
			Blocks: []tg.InputRichBlock{
				{
					Type: "paragraph",
					Text: "Q or W?",
				},
				{
					Type: "paragraph",
					Text: "Ваш ответ: WWW!!!",
				},
				{
					Type: "paragraph",
					Text: "Ваша ставка: ВА-БАНК!",
				},
				{
					Type: "paragraph",
					Text: "Вопрос закрыт. Скоро узнаем, насколько вы были правы.",
				},
			},
		},
	}
	if diff := cmp.Diff(want, msg); diff != "" {
		t.Errorf("request mismatch (-want +got):%s", diff)
	}
}

func TestClosingMessageNoBet(t *testing.T) {
	msg := tg.ClosingMessage(124, 110003, "Cool question", store.Answer{A: "Cool answer"}, store.RealMode)
	want := tg.EditMessageText{
		ChatID:    124,
		MessageID: 110003,
		RichMessage: tg.InputRichMessage{
			Blocks: []tg.InputRichBlock{
				{
					Type: "paragraph",
					Text: "Cool question",
				},
				{
					Type: "paragraph",
					Text: "Ваш ответ: Cool answer",
				},
				{
					Type: "paragraph",
					Text: "Увы, вы не успели сделать ставку. Ваша игра окончена. ",
				},
				{
					Type: "paragraph",
					Text: "Казино прощается с вами!",
				},
			},
		},
	}
	if diff := cmp.Diff(want, msg); diff != "" {
		t.Errorf("request mismatch (-want +got):%s", diff)
	}
}

func TestTutorialClosingMessageNoBet(t *testing.T) {
	msg := tg.ClosingMessage(124, 110003, "Tutorial question", store.Answer{A: "Tutorial answer"}, store.TutorialMode)
	want := tg.EditMessageText{
		ChatID:    124,
		MessageID: 110003,
		RichMessage: tg.InputRichMessage{
			Blocks: []tg.InputRichBlock{
				{
					Type: "paragraph",
					Text: "Tutorial question",
				},
				{
					Type: "paragraph",
					Text: "Ваш ответ: Tutorial answer",
				},
				{
					Type: "paragraph",
					Text: "Увы, вы не успели сделать ставку. Ваша игра была бы окончена, но это всего лишь обучение. ",
				},
				{
					Type: "paragraph",
					Text: "Казино еще не прощается с вами! Ждите следующего вопроса.",
				},
			},
		},
	}
	if diff := cmp.Diff(want, msg); diff != "" {
		t.Errorf("request mismatch (-want +got):%s", diff)
	}
}

func TestClosingMessageNoAnswer(t *testing.T) {
	msg := tg.ClosingMessage(123, 110002, "Question with answer?", store.Answer{}, "")
	want := tg.EditMessageText{
		ChatID:    123,
		MessageID: 110002,
		RichMessage: tg.InputRichMessage{
			Blocks: []tg.InputRichBlock{
				{
					Type: "paragraph",
					Text: "Question with answer?",
				},
				{
					Type: "paragraph",
					Text: "Вопрос закрыт! Вы не успели ответить. По правилам казино вы покидаете игру.",
				},
			},
		},
	}
	if diff := cmp.Diff(want, msg); diff != "" {
		t.Errorf("request mismatch (-want +got):%s", diff)
	}
}

func TestTutorialClosingMessageNoAnswer(t *testing.T) {
	msg := tg.ClosingMessage(123, 110002, "Question with answer?", store.Answer{}, store.TutorialMode)
	want := tg.EditMessageText{
		ChatID:    123,
		MessageID: 110002,
		RichMessage: tg.InputRichMessage{
			Blocks: []tg.InputRichBlock{
				{
					Type: "paragraph",
					Text: "Question with answer?",
				},
				{
					Type: "paragraph",
					Text: "Вопрос закрыт! Вы не успели ответить. По правилам казино вы бы покинули игру, но это всего лишь обучение.",
				},
			},
		},
	}
	if diff := cmp.Diff(want, msg); diff != "" {
		t.Errorf("request mismatch (-want +got):%s", diff)
	}
}

func TestCloseHandlerUsersStoreError(t *testing.T) {
	wantErr := errors.New("users store error")
	spy := &closeStoreSpy{
		usersStoreStub: usersStoreStub{
			err: wantErr,
		},
	}
	h := tg.CloseHandler{
		Store: spy,
	}

	_, err := h.Handle(context.Background(), tg.Update{})
	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Errorf("got err %v; want %v", got, want)
	}
}

func TestCloseHandlerLostUserStoreError(t *testing.T) {
	wantErr := errors.New("lost user store error")
	spy := &closeStoreSpy{
		usersStoreStub: usersStoreStub{
			users: []store.User{
				{},
			},
		},
		lostUserStoreStub: lostUserStoreStub{
			err: wantErr,
		},
	}
	h := tg.CloseHandler{
		Store: spy,
	}

	_, err := h.Handle(context.Background(), tg.Update{})
	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Errorf("got err %v; want %v", got, want)
	}
}

func TestCloseHandlerAnswersStoreError(t *testing.T) {
	wantErr := errors.New("answers store error")
	spy := &closeStoreSpy{
		answersStoreStub: answersStoreStub{
			err: wantErr,
		},
	}
	h := tg.CloseHandler{
		Store: spy,
	}

	_, err := h.Handle(context.Background(), tg.Update{})
	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Errorf("got err %v; want %v", got, want)
	}
}

func TestCloseHandlerQuestionStoreError(t *testing.T) {
	wantErr := errors.New("store error")
	spy := &closeStoreSpy{
		err: wantErr,
	}
	h := tg.CloseHandler{
		Store: spy,
	}

	_, err := h.Handle(context.Background(), tg.Update{})
	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Errorf("got err %v; want %v", got, want)
	}
}

func TestCloseCondition(t *testing.T) {
	h := tg.CloseHandler{}

	ok := h.Condition(tg.Update{
		Message: tg.Message{
			Text: "/closecurrentquestion",
		},
	})

	if got, want := ok, true; got != want {
		t.Errorf("close condition: got %v; want %v", got, want)
	}
}

func TestCloseName(t *testing.T) {
	h := tg.CloseHandler{}

	n := h.Name()

	if got, want := n, "close"; got != want {
		t.Errorf("got name %q; want %q", got, want)
	}
}

type closeStoreSpy struct {
	questionStoreStub
	usersStoreStub
	answersStoreStub
	lostUserStoreStub
	modeStoreStub
	closed bool
	err    error
}

func (s *closeStoreSpy) CloseCurrentQuestion(ctx context.Context) error {
	s.closed = true
	return s.err
}

type answersStoreStub struct {
	questionID string
	answers    []store.Answer
	err        error
}

func (s *answersStoreStub) Answers(ctx context.Context, questionID string) ([]store.Answer, error) {
	s.questionID = questionID
	return s.answers, s.err
}

type lostUserStoreStub struct {
	userIDs []int
	err     error
}

func (s *lostUserStoreStub) SaveLostUser(ctx context.Context, userID int) error {
	s.userIDs = append(s.userIDs, userID)
	return s.err
}

type modeStoreStub struct {
	mode store.Mode
	err  error
}

func (s *modeStoreStub) Mode(ctx context.Context) (store.Mode, error) {
	return s.mode, s.err
}
