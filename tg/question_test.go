package tg_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/synctest"

	"github.com/asahnoln/vote-bot/tg"
	"github.com/google/go-cmp/cmp"
)

// TODO: Separate different parts into different tests?
func TestQuestion(t *testing.T) {
	upd := tg.Update{
		Message: tg.Message{
			Chat: tg.Chat{
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

	destSpy := &destSpy{}
	b := tg.Bot{
		Store: &storeSpy{},
		Dest:  destSpy,
	}
	b.ServeHTTP(w, r)

	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("bot response code: got %v; want %v", got, want)
	}

	want := tg.SendRichMessage{
		ChatID: 123,
		RichMessage: tg.InputRichMessage{
			Blocks: []tg.InputRichBlock{
				{
					Type: "paragraph",
					Text: "X or Y?",
				},
				{
					Type:  "buttons",
					Align: "center",
					Buttons: []tg.RichMessageButton{
						{
							Text:         "X",
							CallbackData: "X",
						},
						{
							Text:         "Y",
							CallbackData: "Y",
						},
					},
				},
			},
		},
	}
	if diff := cmp.Diff(want, destSpy.got); diff != "" {
		t.Errorf("request mismatch (-want +got):%s", diff)
	}
}

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
