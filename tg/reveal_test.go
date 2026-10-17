package tg_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	_ "embed"

	"github.com/asahnoln/vote-bot/store"
	"github.com/asahnoln/vote-bot/tg"
	"github.com/google/go-cmp/cmp"
)

//go:embed testdata/reveal_command_update.json
var revealCommandUpdate []byte

// TODO: Scope to admin
func TestReveal(t *testing.T) {
	storeSpy := &revealStoreSpy{
		answersStoreStub: answersStoreStub{
			answers: []store.Answer{
				{
					UserID:  876,
					IsRight: true,
					Bet:     100,
				},
				{
					UserID:  2345,
					IsRight: false,
					Bet:     200,
				},
				{
					UserID:  15789,
					IsRight: true,
					Bet:     500,
				},
				{
					UserID:  15790,
					IsRight: false,
					Bet:     700,
				},
				// Must be filtered out
				{
					UserID:  15791,
					IsRight: false,
				},
			},
		},
		accounts: map[int]int{
			876:   500,
			2345:  300,
			15789: 0,
			15790: -200,
		},
	}
	h := &tg.RevealHandler{Store: storeSpy}

	upd := tg.Update{}
	json.Unmarshal(revealCommandUpdate, &upd)
	msgs, err := h.Handle(context.Background(), upd)
	if err != nil {
		t.Fatalf("got err %v; want nil", err)
	}

	if got, want := storeSpy.answersStoreStub.questionID, "last-closed-q-99"; got != want {
		t.Errorf("question id for answer: got %q; want %q", got, want)
	}

	if diff := cmp.Diff([]int{876, 2345, 15789, 15790}, storeSpy.userIDs); diff != "" {
		t.Errorf("accounts user IDs mismatch (-want +got):%s", diff)
	}

	if diff := cmp.Diff([]int{15789, 15790}, storeSpy.lostUserStoreStub.userIDs); diff != "" {
		t.Errorf("lost user IDs mismatch (-want +got):%s", diff)
	}

	if got, want := len(msgs), 4; got != want {
		t.Fatalf("request count: got %v; want %v", got, want)
	}

	want := tg.SendRichMessage{
		ChatID: 876,
		RichMessage: tg.InputRichMessage{
			Blocks: []tg.InputRichBlock{
				{
					Type: "paragraph",
					Text: "Да или Нет? Правильный ответ: Да!",
				},
				{
					Type: "paragraph",
					Text: "Вы ответили правильно!",
				},
				{
					Type: "paragraph",
					Text: "Ваш текущий счет - 500.",
				},
			},
		},
	}

	if diff := cmp.Diff(want, msgs[0]); diff != "" {
		t.Errorf("send message mismatch (-want +got):%s", diff)
	}
}

func TestRevealMessage(t *testing.T) {
	msg := tg.RevealMessage(876, "Poo or Foo?", "Foo", true, 500)
	want := tg.SendRichMessage{
		ChatID: 876,
		RichMessage: tg.InputRichMessage{
			Blocks: []tg.InputRichBlock{
				{
					Type: "paragraph",
					Text: "Poo or Foo? Правильный ответ: Foo!",
				},
				{
					Type: "paragraph",
					Text: "Вы ответили правильно!",
				},
				{
					Type: "paragraph",
					Text: "Ваш текущий счет - 500.",
				},
			},
		},
	}
	if diff := cmp.Diff(want, msg); diff != "" {
		t.Errorf("msg mismatch (-want +got):%s", diff)
	}
}

func TestRevealMessageWrongAndBankrupt(t *testing.T) {
	msg := tg.RevealMessage(877, "Saas or Jooj?", "Zheezh", false, 0)
	want := tg.SendRichMessage{
		ChatID: 877,
		RichMessage: tg.InputRichMessage{
			Blocks: []tg.InputRichBlock{
				{
					Type: "paragraph",
					Text: "Saas or Jooj? Правильный ответ: Zheezh!",
				},
				{
					Type: "paragraph",
					Text: "Вы ответили неправильно.",
				},
				{
					Type: "paragraph",
					Text: "Ваш текущий счет - 0.",
				},
				{
					Type: "paragraph",
					Text: "Вы проиграли!",
				},
			},
		},
	}
	if diff := cmp.Diff(want, msg); diff != "" {
		t.Errorf("msg mismatch (-want +got):%s", diff)
	}
}

func TestRevealWrongAnswer(t *testing.T) {
	h := tg.RevealHandler{Store: &revealStoreSpy{
		answersStoreStub: answersStoreStub{
			answers: []store.Answer{{}},
		},
		accounts: map[int]int{
			0: 1,
		},
	}}
	data, err := h.Handle(context.Background(), tg.Update{})
	if err != nil {
		t.Fatalf("got err %v; want nil", err)
	}

	if got, want := data[0].(tg.SendRichMessage).RichMessage.Blocks[1].Text, "Вы ответили неправильно."; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestRevealLastClosedQuestionError(t *testing.T) {
	wantErr := errors.New("last question error")
	h := tg.RevealHandler{Store: &revealStoreSpy{
		lastClosedQuestionErr: wantErr,
	}}

	_, err := h.Handle(context.Background(), tg.Update{})
	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Errorf("got err %q; want %q", got, want)
	}
}

func TestRevealLostUserError(t *testing.T) {
	wantErr := errors.New("lost user error")
	h := tg.RevealHandler{Store: &revealStoreSpy{
		answersStoreStub: answersStoreStub{
			answers: []store.Answer{
				{
					UserID: 100,
				},
			},
		},
		lostUserStoreStub: lostUserStoreStub{
			err: wantErr,
		},
		accounts: map[int]int{
			100: 0,
		},
	}}

	_, err := h.Handle(context.Background(), tg.Update{})
	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Errorf("got err %q; want %q", got, want)
	}
}

func TestRevealAnswersError(t *testing.T) {
	wantErr := errors.New("answers error")
	h := tg.RevealHandler{Store: &revealStoreSpy{
		answersStoreStub: answersStoreStub{
			err: wantErr,
		},
	}}

	_, err := h.Handle(context.Background(), tg.Update{})
	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Errorf("got err %q; want %q", got, want)
	}
}

func TestRevealAccountsError(t *testing.T) {
	wantErr := errors.New("accounts error")
	h := tg.RevealHandler{Store: &revealStoreSpy{
		accountsErr: wantErr,
	}}

	_, err := h.Handle(context.Background(), tg.Update{})
	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Errorf("got err %q; want %q", got, want)
	}
}

func TestRevealCondition(t *testing.T) {
	h := tg.RevealHandler{}

	ok := h.Condition(tg.Update{
		Message: tg.Message{
			Text: "/sendanswer",
		},
	})

	if got, want := ok, true; got != want {
		t.Errorf("reveal true condition: got %v; want %v", got, want)
	}

	ok = h.Condition(tg.Update{})
	if got, want := ok, false; got != want {
		t.Errorf("reveal false condition: got %v; want %v", got, want)
	}
}

type revealStoreSpy struct {
	answersStoreStub
	lostUserStoreStub
	lastClosedQuestionErr error
	userIDs               []int
	accounts              map[int]int
	accountsErr           error
}

func (s *revealStoreSpy) LastClosedQuestion(context.Context) (store.Question, error) {
	return store.Question{
		ID:   "last-closed-q-99",
		Q:    "Да или Нет?",
		A:    0,
		Opts: []string{"Да", "Нет"},
	}, s.lastClosedQuestionErr
}

func (s *revealStoreSpy) Accounts(ctx context.Context, userIDs []int) (map[int]int, error) {
	s.userIDs = userIDs
	return s.accounts, s.accountsErr
}
