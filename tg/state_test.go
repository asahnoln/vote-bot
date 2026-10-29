package tg_test

import (
	"context"
	"errors"
	"testing"

	"github.com/asahnoln/vote-bot/store"
	"github.com/asahnoln/vote-bot/tg"
)

func TestStateWrapperChecksState(t *testing.T) {
	stub := &stateStoreStub{
		state: store.StateAnswer,
	}

	h := &tg.StateWrapper{
		Store:   stub,
		State:   store.StateQuestion,
		Handler: &handlerStub{},
	}
	_, err := h.Handle(context.Background(), tg.Update{})
	if got, want := err, tg.ErrWrongState; !errors.Is(got, want) {
		t.Fatalf("got err %q; want %q", got, want)
	}
}

func TestStateWrapperError(t *testing.T) {
	wantErr := errors.New("state error")
	stub := &stateStoreStub{
		err: wantErr,
	}
	h := tg.StateWrapper{
		Store: stub,
	}
	_, err := h.Handle(context.Background(), tg.Update{})

	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Fatalf("got err %v; want %v", got, want)
	}
}

func TestStateWrapperUsesHandlersMethods(t *testing.T) {
	stub := &stateStoreStub{
		state: store.StateQuestion,
	}

	h := &tg.StateWrapper{Store: stub, State: store.StateQuestion, Handler: &handlerCustomStub{
		conditionFunc: func(u tg.Update) bool {
			return u.Message.Chat.ID == 141
		},
		handleFunc: func(ctx context.Context, u tg.Update) (msgs []tg.Request, err error) {
			msgs = append(msgs, tg.SendRichMessage{
				ChatID: u.Message.Chat.ID,
			})

			return
		},
		handleRespFunc: func(ctx context.Context, a any) error {
			return errors.New("handle resp error: " + a.(string))
		},
		nameFunc: func() string {
			return "test"
		},
	}}
	upd := tg.Update{Message: tg.Message{Chat: tg.Chat{ID: 141}}}
	msgs, err := h.Handle(context.Background(), upd)
	if err != nil {
		t.Fatalf("got err %q; want nil", err)
	}

	if got, want := len(msgs), 1; got != want {
		t.Fatalf("got %d messages; want %d", got, want)
	}
	if got, want := msgs[0].(tg.SendRichMessage).ChatID, 141; got != want {
		t.Errorf("got %v; want %v", got, want)
	}

	ok := h.Condition(upd)
	if got, want := ok, true; got != want {
		t.Errorf("got %v; want %v", got, want)
	}

	err = h.HandleResponse(context.Background(), "cool")
	if got, want := err.Error(), "handle resp error: cool"; got != want {
		t.Errorf("got %v; want %v", got, want)
	}

	if got, want := h.Name(), "test"; got != want {
		t.Errorf("got %v; want %v", got, want)
	}
}

type stateStoreStub struct {
	state store.State
	err   error
}

func (s *stateStoreStub) State(context.Context) (store.State, error) {
	if s.state == "" {
		s.state = store.StateQuestion
	}
	return s.state, s.err
}
