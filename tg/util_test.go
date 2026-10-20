package tg_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/asahnoln/vote-bot/tg"
	"github.com/google/go-cmp/cmp"
)

// * * * *
//
// * * *
// * *
//
// * * *
// * * *
//
// * * * *
// * * *
//
// * * * *
// * * * *
//
// * * *
// * * *
// * * *
//
// * * * *
// * * *
// * * *
//
// * * * *
// * * * *
// * * *
func TestSpreadButtonsCount(t *testing.T) {
	tests := []struct {
		cnt     int
		wantCnt []int
	}{
		{1, []int{1}},
		{4, []int{4}},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d buttons", tt.cnt), func(t *testing.T) {
			opts := make(map[string]bool)
			for i := range tt.cnt {
				opts[strconv.Itoa(i)] = false
			}

			btns := tg.SpreadButtons(opts, "")

			if got, want := len(btns), len(tt.wantCnt); got != want {
				t.Fatalf("btn row count: got %v; want %v", got, want)
			}
			for i, cnt := range tt.wantCnt {
				if got, want := len(btns[i]), cnt; got != want {
					t.Errorf("row %d count: got %v; want %v", i, got, want)
				}
			}
		})
	}
}

func TestSpreadButtonsValues(t *testing.T) {
	btns := tg.SpreadButtons(map[string]bool{
		"opt1": true,
		"opt2": false,
	}, "q-123")

	want := [][]tg.RichMessageButton{
		{
			{Text: "opt1", CallbackData: "opt:q-123:opt1"},
			{Text: "opt2", CallbackData: "opt:q-123:opt2"},
		},
	}

	if diff := cmp.Diff(want, btns); diff != "" {
		t.Errorf("btns mismatch (-want +got)\n:%s", diff)
	}
}
