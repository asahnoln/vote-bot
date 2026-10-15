package tg_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/synctest"

	"github.com/asahnoln/vote-bot/store"
	"github.com/asahnoln/vote-bot/tg"
	"github.com/google/go-cmp/cmp"
)

func TestQuestion(t *testing.T) {
	upd := tg.Update{
		Message: tg.Message{
			Chat: tg.Chat{
				// TODO: Check command scope
				ID: 123,
			},
			Text: "/sendcurrentquestion",
		},
	}

	storeSpy := &struct {
		usersStoreStub
		questionStoreStub
		lastMessageStoreSpy
	}{
		usersStoreStub: usersStoreStub{
			users: []store.User{
				{UserID: 501},
				{UserID: 602},
				{UserID: 703},
			},
		},
	}

	h := &tg.QuestionHandler{Store: storeSpy}
	msgs, err := h.Handle(context.Background(), upd)
	if err != nil {
		t.Fatalf("got err %v; want nil", err)
	}

	if got, want := len(msgs), 3; got != want {
		t.Fatalf("msgs count: got %d; want %d", got, want)
	}
	want := tg.SendRichMessage{
		ChatID: 501,
		RichMessage: tg.InputRichMessage{
			Blocks: []tg.InputRichBlock{
				{
					Type: "paragraph",
					Text: "X or Y?",
				},
				{
					Type:  "buttons",
					Align: "left",
					Buttons: []tg.RichMessageButton{
						{
							Text:         "X",
							CallbackData: "opt:opt-q-99:X",
						},
						{
							Text:         "Y",
							CallbackData: "opt:opt-q-99:Y",
						},
					},
				},
			},
		},
	}
	if diff := cmp.Diff(want, msgs[0]); diff != "" {
		t.Errorf("request mismatch (-want +got):%s", diff)
	}

	if got, want := msgs[1].(tg.SendRichMessage).ChatID, 602; got != want {
		t.Errorf("request 2nd user id: got %v; want %v", got, want)
	}
}

func TestQuestionHandleResponse(t *testing.T) {
	spy := &struct {
		usersStoreStub
		questionStoreStub
		lastMessageStoreSpy
	}{}
	h := tg.QuestionHandler{Store: spy}

	err := h.HandleResponse(context.Background(), tg.Message{
		Chat: tg.Chat{
			ID: 22,
		},
		MessageID: 1853,
	})
	if err != nil {
		t.Fatalf("got err %v; want nil", err)
	}

	if got, want := spy.userID, 22; got != want {
		t.Errorf("last user ID saved: got %v; want %v", got, want)
	}
	if got, want := spy.msgID, 1853; got != want {
		t.Errorf("last msg ID saved: got %v; want %v", got, want)
	}
}

func TestQuestionCondition(t *testing.T) {
	h := tg.QuestionHandler{}

	ok := h.Condition(tg.Update{
		Message: tg.Message{
			Text: "/sendcurrentquestion",
		},
	})
	if got, want := ok, true; got != want {
		t.Errorf("got %v; want %v", got, want)
	}

	ok = h.Condition(tg.Update{})
	if got, want := ok, false; got != want {
		t.Errorf("empty update: got %v; want %v", got, want)
	}
}

// TODO: Check generally in webhook?
func TestQuestionUpdateError(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/webhook", nil)
		w := httptest.NewRecorder()

		errCh := make(chan error)
		go func() {
			err := <-errCh
			if got, want := err.Error(), "update decode: unexpected end of JSON input"; got != want {
				t.Errorf("got err %q; want %q", got, want)
			}
		}()

		b := tg.Bot{
			ErrorChan: errCh,
		}
		b.ServeHTTP(w, r)

		if got, want := w.Code, http.StatusInternalServerError; got != want {
			t.Errorf("bot response code: got %v; want %v", got, want)
		}
	})
}

func TestQuestionStoreError(t *testing.T) {
	wantErr := errors.New("question error")
	h := tg.QuestionHandler{
		Store: &storeSpy{err: wantErr},
	}
	_, err := h.Handle(context.Background(), tg.Update{})

	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Fatalf("got err %v; want %v", got, want)
	}
}

func TestQuestionUsersError(t *testing.T) {
	wantErr := errors.New("users error")
	h := tg.QuestionHandler{
		Store: &storeSpy{usersErr: wantErr},
	}
	_, err := h.Handle(context.Background(), tg.Update{})

	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Fatalf("got err %v; want %v", got, want)
	}
}
