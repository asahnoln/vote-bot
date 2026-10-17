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
// TODO: Make user loose if they missed the question
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
				{UserID: 56, A: "Никто", Bet: 300},
				{UserID: 24, A: "Никто"},
			},
		},
	}
	h := tg.CloseHandler{
		Store: spy,
	}

	msgs, err := h.Handle(context.Background(), tg.Update{
		Message: tg.Message{
			Text: "/closequestion",
		},
	})
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

	want1 := tg.EditMessageText{
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
					Text: "Ваша ставка: 900",
				},
				{
					Type: "paragraph",
					Text: "Вопрос закрыт!",
				},
			},
		},
	}
	if diff := cmp.Diff(want1, msgs[0]); diff != "" {
		t.Errorf("request mismatch (-want +got):%s", diff)
	}

	want2 := tg.EditMessageText{
		ChatID:    56,
		MessageID: 303,
		RichMessage: tg.InputRichMessage{
			Blocks: []tg.InputRichBlock{
				{
					Type: "paragraph",
					Text: "X or Y?",
				},
				{
					Type: "paragraph",
					Text: "Ваш ответ: Никто",
				},
				{
					Type: "paragraph",
					Text: "Ваша ставка: 300",
				},
				{
					Type: "paragraph",
					Text: "Вопрос закрыт!",
				},
			},
		},
	}
	if diff := cmp.Diff(want2, msgs[1]); diff != "" {
		t.Errorf("request 2 mismatch (-want +got):%s", diff)
	}

	want3 := tg.EditMessageText{
		ChatID:    23,
		MessageID: 10002,
		RichMessage: tg.InputRichMessage{
			Blocks: []tg.InputRichBlock{
				{
					Type: "paragraph",
					Text: "X or Y?",
				},
				{
					Type: "paragraph",
					Text: "Вы не успели ответить...",
				},
				{
					Type: "paragraph",
					Text: "Вопрос закрыт!",
				},
			},
		},
	}
	if diff := cmp.Diff(want3, msgs[2]); diff != "" {
		t.Errorf("request 3 mismatch (-want +got):%s", diff)
	}

	want4 := tg.EditMessageText{
		ChatID:    24,
		MessageID: 10003,
		RichMessage: tg.InputRichMessage{
			Blocks: []tg.InputRichBlock{
				{
					Type: "paragraph",
					Text: "X or Y?",
				},
				{
					Type: "paragraph",
					Text: "Ваш ответ: Никто",
				},
				{
					Type: "paragraph",
					Text: "Вы не успели сделать ставку...",
				},
				{
					Type: "paragraph",
					Text: "Ваш ответ не засчитан.",
				},
				{
					Type: "paragraph",
					Text: "Вопрос закрыт!",
				},
			},
		},
	}
	if diff := cmp.Diff(want4, msgs[3]); diff != "" {
		t.Errorf("request 4 mismatch (-want +got):%s", diff)
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
			Text: "/closequestion",
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
	closed bool
	err    error
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

func (s *closeStoreSpy) CloseCurrentQuestion(ctx context.Context) error {
	s.closed = true
	return s.err
}
