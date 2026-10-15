package tg_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/synctest"

	"github.com/asahnoln/vote-bot/store"
	"github.com/asahnoln/vote-bot/tg"
	"github.com/google/go-cmp/cmp"
)

// TODO: Separate different parts into different tests?
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
	body, err := json.Marshal(&upd)
	if err != nil {
		t.Fatalf("marshal err: %v", err)
	}

	r := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	w := httptest.NewRecorder()

	storeSpy := &storeSpy{
		users: []store.User{
			{UserID: 501},
			{UserID: 602},
			{UserID: 703},
		},
	}

	destSpy := &destSpy{
		returnMsg: tg.Message{
			Chat: tg.Chat{
				ID: 501,
			},
			MessageID: 1853,
		},
	}
	b := tg.Bot{
		Store: storeSpy, Dest: destSpy,
	}
	b.ServeHTTP(w, r)

	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("bot response code: got %v; want %v", got, want)
	}

	if got, want := len(destSpy.got), 3; got != want {
		t.Fatalf("dest requests count got %d; want %d", got, want)
	}

	cbData1, _ := json.Marshal(tg.AnswerCallbackData{
		Type:       "opt",
		QuestionID: "opt-q-99",
		Answer:     "X",
	})
	cbData2, _ := json.Marshal(tg.AnswerCallbackData{
		Type:       "opt",
		QuestionID: "opt-q-99",
		Answer:     "Y",
	})
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
							CallbackData: string(cbData1),
						},
						{
							Text:         "Y",
							CallbackData: string(cbData2),
						},
					},
				},
			},
		},
	}
	if diff := cmp.Diff(want, destSpy.got[0]); diff != "" {
		t.Errorf("request mismatch (-want +got):%s", diff)
	}

	if got, want := destSpy.got[1].(tg.SendRichMessage).ChatID, 602; got != want {
		t.Errorf("request 2nd user id: got %v; want %v", got, want)
	}

	if got, want := storeSpy.users[0].LastMessageID, 1853; got != want {
		t.Errorf("last message ID saved: got %v; want %v", got, want)
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

// TODO: Check generally in bot?
func TestQuestionNoStoreError(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader("{}"))
		w := httptest.NewRecorder()

		errCh := make(chan error)
		go func() {
			err := <-errCh
			if got, want := err, tg.ErrStoreNil; !errors.Is(got, want) {
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
	synctest.Test(t, func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader("{}"))
		w := httptest.NewRecorder()

		wantErr := errors.New("store error")
		errCh := make(chan error)
		go func() {
			err := <-errCh
			if got, want := err, wantErr; !errors.Is(got, want) {
				t.Errorf("got err %q; want %q", got, want)
			}
		}()

		b := tg.Bot{
			Store: &storeSpy{
				err: wantErr,
			},
			ErrorChan: errCh,
		}
		b.ServeHTTP(w, r)

		if got, want := w.Code, http.StatusInternalServerError; got != want {
			t.Errorf("bot response code: got %v; want %v", got, want)
		}
	})
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
