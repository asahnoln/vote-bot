package tg_test

import (
	"context"
	"errors"
	"testing"

	"github.com/asahnoln/vote-bot/store"
	"github.com/asahnoln/vote-bot/tg"
	"github.com/google/go-cmp/cmp"
)

func TestStats(t *testing.T) {
	h := tg.StatsHandler{
		Store: &statsStoreStub{},
	}

	msgs, err := h.Handle(context.Background(), tg.Update{})
	if err != nil {
		t.Fatalf("got err %v; want nil", err)
	}

	if got, want := len(msgs), 3; got != want {
		t.Fatalf("msg count: got %v; want %v", got, want)
	}

	want := tg.SendRichMessage{
		ChatID: 500,
		RichMessage: tg.InputRichMessage{
			Blocks: []tg.InputRichBlock{
				{
					Type:      "table",
					Caption:   "Победители",
					IsStriped: true,
					Cells: [][]tg.RichBlockTableCell{
						{
							{
								IsHeader: true,
								Text:     "Имя",
							},
							{
								IsHeader: true,
								Text:     "Кол-во правильных ответов",
							},
							{
								IsHeader: true,
								Text:     "Баланс",
							},
							{
								IsHeader: true,
								Text:     "Среднее время ответа (в секундах)",
							},
						},
						{
							{
								Text:  "Cool name",
								Align: "center",
							},
							{
								Text:  "9",
								Align: "center",
							},
							{
								Text:  "9999$",
								Align: "center",
							},
							{
								Text:  "1.50",
								Align: "center",
							},
						},
						{
							{
								Text:  "Other name",
								Align: "center",
							},
							{
								Text:  "8",
								Align: "center",
							},
							{
								Text:  "9999$",
								Align: "center",
							},
							{
								Text:  "5.50",
								Align: "center",
							},
						},
					},
				},
			},
		},
	}

	if diff := cmp.Diff(want, msgs[0]); diff != "" {
		t.Errorf("send rich message mismatch (-want +got):%s", diff)
	}

	if got, want := msgs[1].(tg.SendRichMessage).ChatID, 501; got != want {
		t.Errorf("2nd msg chat id: got %v; want %v", got, want)
	}
}

func TestStatsRichestAllUsersEror(t *testing.T) {
	wantErr := errors.New("all users error")
	h := tg.StatsHandler{
		Store: &statsStoreStub{
			usersErr: wantErr,
		},
	}

	_, err := h.Handle(context.Background(), tg.Update{})
	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Errorf("got err %q; want %q", got, want)
	}
}

func TestStatsRichestStoreErr(t *testing.T) {
	wantErr := errors.New("store error")
	h := tg.StatsHandler{
		Store: &statsStoreStub{
			err: wantErr,
		},
	}

	_, err := h.Handle(context.Background(), tg.Update{})
	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Errorf("got err %q; want %q", got, want)
	}
}

func TestStatsRichestCondition(t *testing.T) {
	h := tg.StatsHandler{}

	ok := h.Condition(tg.Update{
		Message: tg.Message{
			Text: "/statsrichestusers",
		},
	})

	if got, want := ok, true; got != want {
		t.Errorf("stats richest users condition: got %v; want %v", got, want)
	}

	ok = h.Condition(tg.Update{})
	if got, want := ok, false; got != want {
		t.Errorf("stats richest users false condition: got %v; want %v", got, want)
	}
}

func TestStatsRichestUsersName(t *testing.T) {
	h := tg.StatsHandler{}

	n := h.Name()

	if got, want := n, "stats"; got != want {
		t.Errorf("got name %q; want %q", got, want)
	}
}

type statsStoreStub struct {
	err      error
	usersErr error
}

func (s *statsStoreStub) AllUsers(context.Context) ([]store.User, error) {
	return []store.User{
		{UserID: 500},
		{UserID: 501},
		{UserID: 502},
	}, s.usersErr
}

func (s *statsStoreStub) Stats(context.Context) ([]store.User, error) {
	return []store.User{
		{Name: "Cool name", RightAnswerCnt: 9, Acc: 9999, AvgSpeed: 1.5},
		{Name: "Other name", RightAnswerCnt: 8, Acc: 9999, AvgSpeed: 5.5},
	}, s.err
}
