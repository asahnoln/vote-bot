package tg_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
		accountStoreStub: accountStoreStub{
			accounts: map[int]int{
				876:   500,
				2345:  300,
				15789: 0,
				15790: -200,
			},
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

	if diff := cmp.Diff([]int{876, 2345, 15789, 15790}, storeSpy.accountStoreStub.userIDs); diff != "" {
		t.Errorf("accounts user IDs mismatch (-want +got):%s", diff)
	}

	if diff := cmp.Diff([]int{15789, 15790}, storeSpy.lostUserStoreStub.userIDs); diff != "" {
		t.Errorf("lost user IDs mismatch (-want +got):%s", diff)
	}

	if got, want := len(msgs), 6; got != want {
		t.Fatalf("request count: got %v; want %v", got, want)
	}

	if got, want := msgs[4].(tg.SendRichMessage).ChatID, 10001; got != want {
		t.Errorf("request final admin user id: got %v; want %v", got, want)
	}

	want := tg.SendRichMessage{
		ChatID: 876,
		RichMessage: tg.InputRichMessage{
			Blocks: []tg.InputRichBlock{
				{
					Type: "paragraph",
					Text: "Браво! Правильный ответ. Удача на вашей стороне.",
				},
				{
					Type: "paragraph",
					Text: "Ваш счет — 500$",
				},
			},
		},
	}

	if diff := cmp.Diff(want, msgs[0]); diff != "" {
		t.Errorf("send message mismatch (-want +got):%s", diff)
	}
}

func TestRevealTutorialWontSaveLostUser(t *testing.T) {
	storeSpy := &revealStoreSpy{
		answersStoreStub: answersStoreStub{
			answers: []store.Answer{
				{
					UserID:  15789,
					IsRight: true,
					Bet:     500,
				},
			},
		},
		accountStoreStub: accountStoreStub{
			accounts: map[int]int{
				15789: 0,
			},
		},
		modeStoreStub: modeStoreStub{
			mode: store.TutorialMode,
		},
	}
	h := &tg.RevealHandler{Store: storeSpy}

	upd := tg.Update{}
	json.Unmarshal(revealCommandUpdate, &upd)
	msgs, err := h.Handle(context.Background(), upd)
	if err != nil {
		t.Fatalf("got err %v; want nil", err)
	}

	if got, want := len(msgs), 3; got != want {
		t.Fatalf("request count: got %v; want %v", got, want)
	}
	if got, want := len(storeSpy.lostUserStoreStub.userIDs), 0; got != want {
		t.Fatalf("lost userIDs count: got %v; want %v", got, want)
	}
	if got, want := storeSpy.restoreAcc, 15789; got != want {
		t.Fatalf("restore acc user ID: got %v; want %v", got, want)
	}
}

func TestRevealMessage(t *testing.T) {
	msg := tg.RevealMessage(876, true, 500, "")
	want := tg.SendRichMessage{
		ChatID: 876,
		RichMessage: tg.InputRichMessage{
			Blocks: []tg.InputRichBlock{
				{
					Type: "paragraph",
					Text: "Браво! Правильный ответ. Удача на вашей стороне.",
				},
				{
					Type: "paragraph",
					Text: "Ваш счет — 500$",
				},
			},
		},
	}
	if diff := cmp.Diff(want, msg); diff != "" {
		t.Errorf("msg mismatch (-want +got):%s", diff)
	}
}

func TestRevealMessageWrong(t *testing.T) {
	msg := tg.RevealMessage(878, false, 100, "")
	want := tg.SendRichMessage{
		ChatID: 878,
		RichMessage: tg.InputRichMessage{
			Blocks: []tg.InputRichBlock{
				{
					Type: "paragraph",
					Text: "Не угадали! Вы теряете свою ставку. Посмотрим, удастся ли отыграться.",
				},
				{
					Type: "paragraph",
					Text: "Ваш счет — 100$",
				},
			},
		},
	}
	if diff := cmp.Diff(want, msg); diff != "" {
		t.Errorf("msg mismatch (-want +got):%s", diff)
	}
}

func TestRevealMessageWrongAndBankrupt(t *testing.T) {
	tests := []struct {
		acc int
	}{{0}, {-100}}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("acc %d$", tt.acc), func(t *testing.T) {
			msg := tg.RevealMessage(877, false, tt.acc, store.RealMode)
			want := tg.SendRichMessage{
				ChatID: 877,
				RichMessage: tg.InputRichMessage{
					Blocks: []tg.InputRichBlock{
						{
							Type: "paragraph",
							Text: "Неверный ответ.",
						},
						{
							Type: "paragraph",
							Text: "Ваш счет — 0$",
						},
						{
							Type: "paragraph",
							Text: "Вы банкрот. Казино прощается с вами!",
						},
					},
				},
			}
			if diff := cmp.Diff(want, msg); diff != "" {
				t.Errorf("msg mismatch (-want +got):%s", diff)
			}
		})
	}
}

func TestRevealMessageBankruptTutorial(t *testing.T) {
	msg := tg.RevealMessage(877, false, 0, store.TutorialMode)
	want := tg.SendRichMessage{
		ChatID: 877,
		RichMessage: tg.InputRichMessage{
			Blocks: []tg.InputRichBlock{
				{
					Type: "paragraph",
					Text: "Неверный ответ.",
				},
				{
					Type: "paragraph",
					Text: "Ваш счет — 0$",
				},
				{
					Type: "paragraph",
					Text: "Вы банкрот. Казино бы попрощалось с вами, но это всего лишь обучение. Мы снова добавляем вам 1000$ на образовательные нужды.",
				},
			},
		},
	}
	if diff := cmp.Diff(want, msg); diff != "" {
		t.Errorf("msg mismatch (-want +got):%s", diff)
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
		accountStoreStub: accountStoreStub{
			accounts: map[int]int{
				100: 0,
			},
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
		accountStoreStub: accountStoreStub{
			err: wantErr,
		},
	}}

	_, err := h.Handle(context.Background(), tg.Update{})
	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Errorf("got err %q; want %q", got, want)
	}
}

func TestRevealHandlerModeStoreError(t *testing.T) {
	wantErr := errors.New("mode error")
	spy := &revealStoreSpy{
		modeStoreStub: modeStoreStub{
			err: wantErr,
		},
	}
	h := tg.RevealHandler{
		Store: spy,
	}

	_, err := h.Handle(context.Background(), tg.Update{})
	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Errorf("got err %v; want %v", got, want)
	}
}

func TestRevealHandlerRestoreAccError(t *testing.T) {
	wantErr := errors.New("restore acc error")
	spy := &revealStoreSpy{
		answersStoreStub: answersStoreStub{
			answers: []store.Answer{
				{
					UserID: 101,
				},
			},
		},
		accountStoreStub: accountStoreStub{
			accounts: map[int]int{
				101: 0,
			},
		},
		modeStoreStub: modeStoreStub{
			mode: store.TutorialMode,
		},
		restoreAccErr: wantErr,
	}
	h := tg.RevealHandler{
		Store: spy,
	}

	_, err := h.Handle(context.Background(), tg.Update{})
	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Errorf("got err %v; want %v", got, want)
	}
}

func TestRevealHandlerAdminsError(t *testing.T) {
	wantErr := errors.New("admins error")
	spy := &revealStoreSpy{
		answersStoreStub: answersStoreStub{
			answers: []store.Answer{
				{
					UserID: 101,
				},
			},
		},
		adminsStoreStub: adminsStoreStub{
			err: wantErr,
		},
	}
	h := tg.RevealHandler{
		Store: spy,
	}

	_, err := h.Handle(context.Background(), tg.Update{})
	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Errorf("got err %v; want %v", got, want)
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
	accountStoreStub
	modeStoreStub
	adminsStoreStub
	stateStoreSpy
	lastClosedQuestionErr error
	restoreAcc            int
	restoreAccErr         error
}

func (s *revealStoreSpy) LastClosedQuestion(context.Context) (store.Question, error) {
	return store.Question{
		ID:   "last-closed-q-99",
		Q:    "Да или Нет?",
		Opts: map[string]bool{"Да": true, "Нет": false},
	}, s.lastClosedQuestionErr
}

func (s *revealStoreSpy) RestoreAcc(ctx context.Context, userID int) error {
	s.restoreAcc = userID
	return s.restoreAccErr
}

type accountStoreStub struct {
	userIDs  []int
	accounts map[int]int
	err      error
}

func (s *accountStoreStub) Accounts(ctx context.Context, userIDs []int) (map[int]int, error) {
	s.userIDs = userIDs
	return s.accounts, s.err
}
