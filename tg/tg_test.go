package tg_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/asahnoln/vote-bot/tg"
	"github.com/google/go-cmp/cmp"
)

// TODO: Check webhook target

func TestQuestion(t *testing.T) {
	got := tg.SendRichMessage{}
	tgSrvStubCalled := false
	tgSrvStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tgSrvStubCalled = true

		if got, want := r.Method, http.MethodPost; got != want {
			t.Errorf("tg srv request method: got %q; want %q", got, want)
		}

		if got, want := r.Header.Get("Content-Type"), "application/json"; got != want {
			t.Errorf("tg srv request method: got %q; want %q", got, want)
		}

		if got, want := r.URL.Path, "/sendRichMessage"; got != want {
			t.Errorf("tg srv request path: got %q; want %q", got, want)
		}

		err := json.NewDecoder(r.Body).Decode(&got)
		if err != nil {
			t.Errorf("tg srv request unmarshal error: %v", err)
		}

		defer r.Body.Close()
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
							Text: "X",
						},
						{
							Text: "Y",
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

// TODO: Check for store to be passed

func TestQuestionUpdateError(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/webhook", nil)
	w := httptest.NewRecorder()

	b := tg.Bot{}
	b.ServeHTTP(w, r)

	if got, want := w.Code, http.StatusInternalServerError; got != want {
		t.Errorf("bot response code: got %v; want %v", got, want)
	}
}

type stubStore struct{}

func (s *stubStore) CurrentQuestion() (tg.Question, error) {
	return tg.Question{
		Q:    "X or Y?",
		Opts: []string{"X", "Y"},
	}, nil
}
