package tg_test

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/synctest"

	"github.com/asahnoln/vote-bot/tg"
	"github.com/google/go-cmp/cmp"
)

//go:embed testdata/answer_callback_update.json
var answerCallbackUpdate []byte

func TestAnswer(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(answerCallbackUpdate))
	w := httptest.NewRecorder()

	storeSpy := &storeSpy{}
	destSpy := &destSpy{}
	b := tg.Bot{
		Store: storeSpy,
		Dest:  destSpy,
	}
	b.ServeHTTP(w, r)

	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("answer status: got %v; want %v", got, want)
	}

	want := tg.EditMessageText{
		ChatID:    45030020,
		MessageID: 79,
		RichMessage: tg.InputRichMessage{
			Blocks: []tg.InputRichBlock{
				{
					Type: "paragraph",
					Text: "Выберите ставку:",
				},
				{
					Type:  "buttons",
					Align: "left",
					Buttons: []tg.RichMessageButton{
						{
							Text:         "100",
							CallbackData: "bet:opt-q-67:100",
						},
						{
							Text:         "200",
							CallbackData: "bet:opt-q-67:200",
						},
						{
							Text:         "500",
							CallbackData: "bet:opt-q-67:500",
						},
					},
				},
			},
		},
	}
	if diff := cmp.Diff(want, destSpy.got); diff != "" {
		t.Errorf("edit message text mismatch (-want +got):%s", diff)
	}

	if got, want := storeSpy.userID, 45030020; got != want {
		t.Errorf("save answer: got %v; want %v", got, want)
	}
	if got, want := storeSpy.answer, "Ты"; got != want {
		t.Errorf("save answer: got %v; want %v", got, want)
	}
}

func TestAnswerSaveError(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(answerCallbackUpdate))
		w := httptest.NewRecorder()

		wantErr := errors.New("save error")

		errCh := make(chan error)
		go func() {
			err := <-errCh
			if got, want := err, wantErr; !errors.Is(got, want) {
				t.Fatalf("got err %v; want %v", got, want)
			}
		}()

		storeSpy := &storeSpy{
			err: wantErr,
		}
		b := tg.Bot{
			Store:     storeSpy,
			ErrorChan: errCh,
		}
		b.ServeHTTP(w, r)

		if got, want := w.Code, http.StatusInternalServerError; got != want {
			t.Errorf("answer status: got %v; want %v", got, want)
		}
	})
}

//go:embed testdata/bet_callback_update.json
var betCallbackUpdate []byte

func TestBet(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(betCallbackUpdate))
	w := httptest.NewRecorder()

	storeSpy := &storeSpy{}
	destSpy := &destSpy{}
	b := tg.Bot{
		Store: storeSpy,
		Dest:  destSpy,
	}
	b.ServeHTTP(w, r)

	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("bet status: got %v; want %v", got, want)
	}

	want := tg.EditMessageText{
		ChatID:    45030021,
		MessageID: 105,
		RichMessage: tg.InputRichMessage{
			Blocks: []tg.InputRichBlock{
				{
					Type: "paragraph",
					// TODO: Подумать над ответом
					Text: "Ваш ответ и ваша ставка приняты!",
				},
			},
		},
	}
	if diff := cmp.Diff(want, destSpy.got); diff != "" {
		t.Errorf("edit message text mismatch (-want +got):%s", diff)
	}

	if got, want := storeSpy.userID, 45030021; got != want {
		t.Errorf("save bet: got %v; want %v", got, want)
	}
	if got, want := storeSpy.bet, 500; got != want {
		t.Errorf("save bet: got %v; want %v", got, want)
	}
}

func TestBetSaveError(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(betCallbackUpdate))
		w := httptest.NewRecorder()

		wantErr := errors.New("save error")
		errCh := make(chan error)
		go func() {
			err := <-errCh
			if got, want := err, wantErr; !errors.Is(got, want) {
				t.Fatalf("got err %v; want %v", got, want)
			}
		}()

		storeSpy := &storeSpy{
			err: wantErr,
		}
		b := tg.Bot{
			Store:     storeSpy,
			ErrorChan: errCh,
		}
		b.ServeHTTP(w, r)

		if got, want := w.Code, http.StatusInternalServerError; got != want {
			t.Errorf("bet status: got %v; want %v", got, want)
		}
	})
}

//go:embed testdata/bet_callback_update_wrong_bet.json
var betCallbackUpdateWrongBet []byte

func TestBetWrongBetFormatError(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(betCallbackUpdateWrongBet))
		w := httptest.NewRecorder()

		errCh := make(chan error)
		go func() {
			err := <-errCh
			if got, want := err, tg.ErrBetIsNotANumber; !errors.Is(got, want) {
				t.Fatalf("got err %v; want %v", got, want)
			}
		}()

		b := tg.Bot{
			ErrorChan: errCh,
		}
		b.ServeHTTP(w, r)

		if got, want := w.Code, http.StatusInternalServerError; got != want {
			t.Errorf("bet status: got %v; want %v", got, want)
		}
	})
}

//go:embed testdata/bet_callback_update_wrong_data.json
var betCallbackUpdateWrongData []byte

func TestBetWrongDataFormatError(t *testing.T) {
	h := tg.BetHandler{}

	_, err := h.Handle(context.Background(), tg.Update{
		CallbackQuery: &tg.CallbackQuery{
			Data: "bet: this is very wrong data ",
		},
	})
	if got, want := err, tg.ErrWrongCallbackData; !errors.Is(got, want) {
		t.Fatalf("got err %v; want %v", got, want)
	}
}

type destSpy struct {
	got any
	err error
}

func (d *destSpy) Send(ctx context.Context, data tg.URLer) error {
	d.got = data
	return d.err
}
