package tg_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/synctest"

	"github.com/asahnoln/vote-bot/store"
	"github.com/asahnoln/vote-bot/tg"
)

func init() {
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
}

// TODO: Check if update is correct for every type of action, have a default action as well
func TestDefaultNotFound(t *testing.T) {
	r := httptest.NewRequest("", "/", nil)
	w := httptest.NewRecorder()

	b := tg.Bot{}
	b.ServeHTTP(w, r)

	if got, want := w.Code, http.StatusNotFound; got != want {
		t.Errorf("bot response code: got %v; want %v", got, want)
	}
}

func TestBotUpdateError(t *testing.T) {
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

		// if got, want := w.Code, http.StatusInternalServerError; got != want {
		// 	t.Errorf("bot response code: got %v; want %v", got, want)
		// }
	})
}

func TestBotDestError(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := httptest.NewRequest("", "/webhook", strings.NewReader("{}"))
		w := httptest.NewRecorder()

		errCh := make(chan error)
		wantErr := errors.New("dest error")
		go func() {
			err := <-errCh
			if got, want := err, wantErr; !errors.Is(got, want) {
				t.Errorf("got err %v; want %v", got, want)
			}
		}()

		b := tg.Bot{
			Dest:      &destSpy{err: wantErr},
			ErrorChan: errCh,
		}
		b.Handle(&handlerStub{})
		b.ServeHTTP(w, r)

		if got, want := w.Code, http.StatusOK; got != want {
			t.Errorf("bot response code: got %v; want %v", got, want)
		}
	})
}

func TestBotHandlers(t *testing.T) {
	tests := []struct {
		updText    string
		wantChatID int
	}{
		{"handler1", 10},
		{"handler2", 20},
	}

	handlerStubBase := &handlerStub{}
	handlerStub1 := &handlerCustomStub{
		conditionFunc: func(u tg.Update) bool {
			return u.Message.Text == "handler1"
		},
		handleFunc: func(ctx context.Context, u tg.Update) (msgs []tg.Request, err error) {
			msgs = append(msgs, tg.SendRichMessage{
				ChatID: 10,
			})
			return
		},
		handleRespFunc: handlerStubBase.HandleResponse,
		nameFunc:       handlerStubBase.Name,
	}
	handlerStub2 := &handlerCustomStub{
		conditionFunc: func(u tg.Update) bool {
			return u.Message.Text == "handler2"
		},
		handleFunc: func(ctx context.Context, u tg.Update) (msgs []tg.Request, err error) {
			msgs = append(msgs, tg.SendRichMessage{
				ChatID: 20,
			})
			return
		},
		handleRespFunc: handlerStubBase.HandleResponse,
		nameFunc:       handlerStubBase.Name,
	}

	b := tg.Bot{}
	b.Handle(handlerStub1)
	b.Handle(handlerStub2)

	for _, tt := range tests {
		t.Run(tt.updText, func(t *testing.T) {
			upd := tg.Update{
				Message: tg.Message{
					Text: tt.updText,
				},
			}

			body, err := json.Marshal(&upd)
			if err != nil {
				t.Fatalf("marshal err: %v", err)
			}

			r := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
			w := httptest.NewRecorder()

			dest := &destSpy{}
			b.Dest = dest
			b.ServeHTTP(w, r)

			if got, want := w.Code, http.StatusOK; got != want {
				t.Errorf("bot response code: got %v; want %v", got, want)
			}

			if got, want := dest.got[0].(tg.SendRichMessage).ChatID, tt.wantChatID; got != want {
				t.Errorf("sent msg with chat ID: got %v; want %v", got, want)
			}
		})
	}
}

func TestBotHandler(t *testing.T) {
	upd := tg.Update{
		Message: tg.Message{
			Text: "specialText",
		},
	}
	body, err := json.Marshal(&upd)
	if err != nil {
		t.Fatalf("marshal err: %v", err)
	}

	r := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	w := httptest.NewRecorder()

	destSpy := &destSpy{
		returnMsg: tg.Message{
			Chat: tg.Chat{
				ID: 501,
			},
			MessageID: 1853,
		},
	}
	b := tg.Bot{
		Dest: destSpy,
	}

	gotM := tg.Message{}
	b.Handle(&handlerCustomStub{
		conditionFunc: func(u tg.Update) bool {
			return u.Message.Text == "specialText"
		},
		handleFunc: func(ctx context.Context, u tg.Update) (msgs []tg.Request, err error) {
			msgs = append(msgs, tg.SendRichMessage{
				ChatID: 96712,
			})
			msgs = append(msgs, tg.EditMessageText{
				ChatID: 602,
			})
			msgs = append(msgs, tg.SendRichMessage{})
			return
		},
		handleRespFunc: func(ctx context.Context, m any) error {
			gotM = m.(tg.Message)
			return nil
		},
		nameFunc: func() string {
			return "customStub"
		},
	})
	b.ServeHTTP(w, r)

	if got, want := len(destSpy.got), 3; got != want {
		t.Fatalf("dest requests count got %d; want %d", got, want)
	}

	if got, want := destSpy.got[0].(tg.SendRichMessage).ChatID, 96712; got != want {
		t.Errorf("dest request msg chatID: got %d; want %d", got, want)
	}

	if got, want := destSpy.got[1].(tg.EditMessageText).ChatID, 602; got != want {
		t.Errorf("request 2nd user id: got %v; want %v", got, want)
	}

	if got, want := gotM.MessageID, 1853; got != want {
		t.Errorf("last message handled: got %v; want %v", got, want)
	}
}

func TestBotMessagesSentWhenDestErrorsAnytime(t *testing.T) {
	upd := tg.Update{
		Message: tg.Message{
			Text: "specialText",
		},
	}
	body, err := json.Marshal(&upd)
	if err != nil {
		t.Fatalf("marshal err: %v", err)
	}

	r := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	w := httptest.NewRecorder()

	destSpy := &destSpy{
		err: errors.New("dest error"),
	}
	b := tg.Bot{
		Dest: destSpy,
	}

	b.Handle(&handlerCustomStub{
		conditionFunc: func(u tg.Update) bool {
			return u.Message.Text == "specialText"
		},
		handleFunc: func(ctx context.Context, u tg.Update) (msgs []tg.Request, err error) {
			msgs = append(msgs, tg.SendRichMessage{
				ChatID: 96712,
			})
			msgs = append(msgs, tg.EditMessageText{
				ChatID: 602,
			})
			msgs = append(msgs, tg.DeleteMessage{})
			return
		},
		handleRespFunc: func(ctx context.Context, m any) error {
			return nil
		},
		nameFunc: func() string {
			return "customStub"
		},
	})
	b.ServeHTTP(w, r)

	if got, want := len(destSpy.got), 3; got != want {
		t.Fatalf("dest requests count got %d; want %d", got, want)
	}
}

type usersStoreStub struct {
	users []store.User
	err   error
}

func (s *usersStoreStub) Users(context.Context) ([]store.User, error) {
	return s.users, s.err
}

type questionStoreStub struct {
	err error
}

func (s *questionStoreStub) CurrentQuestion(context.Context) (store.Question, error) {
	return store.Question{
		ID:   "opt-q-99",
		Q:    "X or Y?",
		Opts: map[string]bool{"X": false, "Y": true},
	}, s.err
}

type lastMessageStoreSpy struct {
	userID int
	msgID  int
	err    error
}

func (s *lastMessageStoreSpy) SaveLastMessageID(ctx context.Context, userID int, msgID int) error {
	s.userID = userID
	s.msgID = msgID

	return s.err
}

type handlerStub struct{}

func (s *handlerStub) Condition(tg.Update) bool {
	return true
}

func (s *handlerStub) Handle(context.Context, tg.Update) (msgs []tg.Request, err error) {
	msgs = append(msgs, tg.EditMessageText{})
	return
}

func (s *handlerStub) HandleResponse(context.Context, any) error {
	return nil
}

func (s *handlerStub) Name() string {
	return "stub"
}

type handlerCustomStub struct {
	conditionFunc  func(tg.Update) bool
	handleFunc     func(context.Context, tg.Update) ([]tg.Request, error)
	handleRespFunc func(context.Context, any) error
	nameFunc       func() string
}

func (s *handlerCustomStub) Condition(upd tg.Update) bool {
	return s.conditionFunc(upd)
}

func (s *handlerCustomStub) Handle(ctx context.Context, upd tg.Update) (msgs []tg.Request, err error) {
	return s.handleFunc(ctx, upd)
}

func (s *handlerCustomStub) HandleResponse(ctx context.Context, m any) error {
	return s.handleRespFunc(ctx, m)
}

func (s *handlerCustomStub) Name() string {
	return s.nameFunc()
}

type destSpy struct {
	got       []any
	err       error
	returnMsg tg.Message
}

func (d *destSpy) Send(ctx context.Context, data tg.Request) (m any, err error) {
	d.got = append(d.got, data)
	return d.returnMsg, d.err
}

type textStub struct{}

func (t *textStub) TextWelcomeToTheGame() string {
	return "TextWelcomeToTheGame"
}

func (t *textStub) TextMissedFirstQuestion() string {
	return "TextMissedFirstQuestion"
}

func (t *textStub) TextAlreadyJoinedTheGame() string {
	return "TextAlreadyJoinedTheGame"
}
