package tg_test

import (
	"context"
	"errors"
	"testing"

	"github.com/asahnoln/vote-bot/store"
	"github.com/asahnoln/vote-bot/tg"
	"github.com/google/go-cmp/cmp"
)

// TODO: Limit to admins

// TODO: Check command scope
func TestQuestion(t *testing.T) {
	storeSpy := &struct {
		usersStoreStub
		questionStoreStub
		lastMessageStoreSpy
		questionSentAtStoreSpy
		adminsStoreStub
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
	msgs, err := h.Handle(context.Background(), tg.Update{
		Message: tg.Message{
			Chat: tg.Chat{
				ID: 123,
			},
		},
	})
	if err != nil {
		t.Fatalf("got err %v; want nil", err)
	}

	if got, want := len(msgs), 5; got != want {
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

	if got, want := storeSpy.questionSentAtStoreSpy.called, true; got != want {
		t.Errorf("question save sent at called: got %v; want %v", got, want)
	}

	if got, want := msgs[3].(tg.SendRichMessage).ChatID, 10001; got != want {
		t.Errorf("request final admin user id: got %v; want %v", got, want)
	}
}

func TestQuestionHandleResponse(t *testing.T) {
	spy := &struct {
		usersStoreStub
		questionStoreStub
		lastMessageStoreSpy
		questionSentAtStoreSpy
		adminsStoreStub
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

func TestQuestionHandleResponseDontSaveAdminMessage(t *testing.T) {
	spy := &struct {
		usersStoreStub
		questionStoreStub
		lastMessageStoreSpy
		questionSentAtStoreSpy
		adminsStoreStub
	}{}
	h := tg.QuestionHandler{Store: spy}

	err := h.HandleResponse(context.Background(), tg.Message{
		Chat: tg.Chat{
			ID: 22,
		},
		MessageID: 1853,
		RichMessage: tg.InputRichMessage{
			Blocks: []tg.InputRichBlock{
				{
					Text: "Сообщение только для админов",
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("got err %v; want nil", err)
	}

	if got, want := spy.userID, 0; got != want {
		t.Errorf("last user ID saved: got %v; want %v", got, want)
	}
	if got, want := spy.msgID, 0; got != want {
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

func TestQuestionStoreError(t *testing.T) {
	wantErr := errors.New("question error")
	stub := &struct {
		usersStoreStub
		questionStoreStub
		lastMessageStoreSpy
		questionSentAtStoreSpy
		adminsStoreStub
	}{
		questionStoreStub: questionStoreStub{
			err: wantErr,
		},
	}
	h := tg.QuestionHandler{
		Store: stub,
	}
	_, err := h.Handle(context.Background(), tg.Update{})

	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Fatalf("got err %v; want %v", got, want)
	}
}

func TestQuestionUsersError(t *testing.T) {
	wantErr := errors.New("users error")
	stub := &struct {
		usersStoreStub
		questionStoreStub
		lastMessageStoreSpy
		questionSentAtStoreSpy
		adminsStoreStub
	}{
		usersStoreStub: usersStoreStub{
			err: wantErr,
		},
	}
	h := tg.QuestionHandler{
		Store: stub,
	}
	_, err := h.Handle(context.Background(), tg.Update{})

	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Fatalf("got err %v; want %v", got, want)
	}
}

func TestQuestionSaveSentAtError(t *testing.T) {
	wantErr := errors.New("question sent at error")
	stub := &struct {
		usersStoreStub
		questionStoreStub
		lastMessageStoreSpy
		questionSentAtStoreSpy
		adminsStoreStub
	}{
		questionSentAtStoreSpy: questionSentAtStoreSpy{
			err: wantErr,
		},
	}
	h := tg.QuestionHandler{
		Store: stub,
	}
	_, err := h.Handle(context.Background(), tg.Update{})

	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Fatalf("got err %v; want %v", got, want)
	}
}

func TestQuestionAdminsError(t *testing.T) {
	wantErr := errors.New("admins error")
	stub := &struct {
		usersStoreStub
		questionStoreStub
		lastMessageStoreSpy
		questionSentAtStoreSpy
		adminsStoreStub
	}{
		adminsStoreStub: adminsStoreStub{
			err: wantErr,
		},
	}
	h := tg.QuestionHandler{
		Store: stub,
	}
	_, err := h.Handle(context.Background(), tg.Update{})

	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Fatalf("got err %v; want %v", got, want)
	}
}

type questionSentAtStoreSpy struct {
	called bool
	err    error
}

func (s *questionSentAtStoreSpy) SaveCurrentQuestionSentAt(context.Context) error {
	s.called = true
	return s.err
}
