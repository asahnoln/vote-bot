package tg_test

import (
	"context"
	"errors"
	"testing"

	"github.com/asahnoln/vote-bot/store"
	"github.com/asahnoln/vote-bot/tg"
	"github.com/google/go-cmp/cmp"
)

func TestStatsRichest(t *testing.T) {
	h := tg.StatsRichestHandler{
		Store: &richestStoreStub{},
	}

	msgs, err := h.Handle(context.Background(), tg.Update{})
	if err != nil {
		t.Fatalf("got err %v; want nil", err)
	}

	want := tg.SendRichMessage{
		ChatID: 500,
		RichMessage: tg.InputRichMessage{
			Blocks: []tg.InputRichBlock{
				{
					Type:      "table",
					Caption:   "Самые богатые",
					IsStriped: true,
					Cells: [][]tg.RichBlockTableCell{
						{
							{
								Text:  "Cool name",
								Align: "center",
							},
							{
								Text:  "9999$",
								Align: "center",
							},
						},
						{
							{
								Text:  "Other name",
								Align: "center",
							},
							{
								Text:  "9999$",
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
}

func TestStatsRichestStoreErr(t *testing.T) {
	wantErr := errors.New("store error")
	h := tg.StatsRichestHandler{
		Store: &richestStoreStub{
			err: wantErr,
		},
	}

	_, err := h.Handle(context.Background(), tg.Update{})
	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Errorf("got err %q; want %q", got, want)
	}
}

type richestStoreStub struct {
	err error
}

func (s *richestStoreStub) RichestUsers(context.Context) ([]store.User, error) {
	return []store.User{
		{Name: "Cool name", Acc: 9999},
		{Name: "Other name", Acc: 9999},
	}, s.err
}
