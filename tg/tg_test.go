package tg_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/synctest"

	"github.com/asahnoln/vote-bot/tg"
	"github.com/google/go-cmp/cmp"
)

// TODO: Check webhook target

// TODO: Separate different parts into different tests?
func TestQuestion(t *testing.T) {
	got := tg.SendRichMessage{}
	tgSrvStubCalled := false
	tgSrvStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tgSrvStubCalled = true

		if got, want := r.Method, http.MethodPost; got != want {
			t.Errorf("tg srv request method: got %q; want %q", got, want)
		}

		if got, want := r.Header.Get("Content-Type"), "application/json"; got != want {
			t.Errorf("tg srv request Content-Type: got %q; want %q", got, want)
		}

		if got, want := r.URL.Path, "/sendRichMessage"; got != want {
			t.Errorf("tg srv request path: got %q; want %q", got, want)
		}

		err := json.NewDecoder(r.Body).Decode(&got)
		if err != nil {
			t.Errorf("tg srv request unmarshal error: %v", err)
		}
		defer r.Body.Close()

		resp := tg.Response{
			OK: true,
		}

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			t.Fatalf("tg srv resp err: %v", err)
		}
	}))
	defer tgSrvStub.Close()

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

	b := tg.Bot{
		Store: &stubStore{},
		URL:   tgSrvStub.URL,
	}
	b.ServeHTTP(w, r)

	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("bot response code: got %v; want %v", got, want)
	}

	if got, want := tgSrvStubCalled, true; got != want {
		t.Fatalf("tg srv stub called: got %v; want %v", got, want)
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
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("request mismatch (-want +got):%s", diff)
	}
}

func TestTgSrvNotOK(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		tgSrvStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := tg.Response{
				OK:          false,
				Description: "something is not ok",
			}

			err := json.NewEncoder(w).Encode(resp)
			if err != nil {
				t.Fatalf("tg srv resp err: %v", err)
			}
		}))
		defer tgSrvStub.Close()

		r := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader("{}"))
		w := httptest.NewRecorder()

		errCh := make(chan error)
		go func() {
			err := <-errCh
			if got, want := err.Error(), "tg srv not ok: something is not ok"; got != want {
				t.Errorf("got err %q; want %q", got, want)
			}
		}()

		b := tg.Bot{
			Store:     &stubStore{},
			URL:       tgSrvStub.URL,
			ErrorChan: errCh,
		}
		b.ServeHTTP(w, r)

		if got, want := w.Code, http.StatusServiceUnavailable; got != want {
			t.Errorf("bot response code: got %v; want %v", got, want)
		}
	})
}

func TestTgSrvNotOKWrongJSON(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		tgSrvStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(``))
		}))
		defer tgSrvStub.Close()

		r := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader("{}"))
		w := httptest.NewRecorder()

		errCh := make(chan error)
		go func() {
			err := <-errCh
			if got, want := err, io.EOF; !errors.Is(got, want) {
				t.Errorf("got err %v; want %q", got, want)
			}
		}()

		b := tg.Bot{
			Store:     &stubStore{},
			URL:       tgSrvStub.URL,
			ErrorChan: errCh,
		}
		b.ServeHTTP(w, r)

		if got, want := w.Code, http.StatusInternalServerError; got != want {
			t.Errorf("bot response code: got %v; want %v", got, want)
		}
	})
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
			Store: &stubStore{
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

func TestQuestionTgSrvRequestError(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader("{}"))
		w := httptest.NewRecorder()

		errCh := make(chan error)
		go func() {
			err := <-errCh
			if err == nil {
				t.Error("got err nil; want not nil")
			}
		}()

		b := tg.Bot{
			Store:     &stubStore{},
			ErrorChan: errCh,
		}
		b.ServeHTTP(w, r)

		if got, want := w.Code, http.StatusInternalServerError; got != want {
			t.Errorf("bot response code: got %v; want %v", got, want)
		}
	})
}

type stubStore struct {
	err error
}

func (s *stubStore) CurrentQuestion(context.Context) (tg.Question, error) {
	return tg.Question{
		Q:    "X or Y?",
		Opts: []string{"X", "Y"},
	}, s.err
}
