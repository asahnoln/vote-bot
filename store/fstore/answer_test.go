package fstore_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/asahnoln/vote-bot/store"
	"github.com/asahnoln/vote-bot/store/fstore"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestSaveAnswer(t *testing.T) {
	ctx := context.Background()
	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	c.Collection("answers").Doc("52:test-q-67").Delete(ctx)
	c.Collection(fstore.QuestionsCollection).Doc("test-q-67").Set(ctx, store.Question{
		Q:     "52 or 42?",
		Opts:  map[string]bool{"wrong": false, "right1": true, "right2": true},
		Order: 0,
	})

	f := fstore.New(c)

	tests := []struct {
		answer string
		want   bool
	}{
		{"right1", true},
		{"wrong", false},
		{"right2", true},
	}

	for _, tt := range tests {
		t.Run(tt.answer, func(t *testing.T) {
			err := f.SaveAnswer(ctx, 52, "test-q-67", tt.answer)
			if err != nil {
				t.Fatalf("fstore save answer: got err %q; want nil", err)
			}

			d, err := c.Collection("answers").Doc("52:test-q-67").Get(ctx)
			if err != nil {
				t.Fatalf("get doc err: %v", err)
			}

			now := time.Now()
			want := map[string]any{
				"question_id": "test-q-67",
				"user_id":     int64(52),
				"a":           tt.answer,
				"is_right":    tt.want,
				"answered_at": now,
			}
			if diff := cmp.Diff(want, d.Data(), cmpopts.EquateApproxTime(time.Second)); diff != "" {
				t.Errorf("answer mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSaveAnswerNoQuestion(t *testing.T) {
	ctx := context.Background()
	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	f := fstore.New(c)

	err := f.SaveAnswer(ctx, 999, "test-q-9999", "")
	if err == nil {
		t.Fatal("fstore save answer: got err nil; want not nil")
	}
}

func TestSaveBet(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	tests := []struct {
		bet       int
		wantBet   int64
		wantAllIn bool
	}{
		{500, 500, false},
		{-100, 1000, true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("bet %d", tt.bet), func(t *testing.T) {
			c.Collection("answers").Doc("234:test-q-111").Set(ctx, map[string]any{
				"user_id": 234,
				"a":       "has answer",
			})

			f := fstore.New(c)

			err := f.SaveBet(ctx, 234, "test-q-111", tt.bet)
			if err != nil {
				t.Fatalf("fstore save bet: got err %q; want nil", err)
			}

			d, err := c.Collection("answers").Doc("234:test-q-111").Get(ctx)
			if err != nil {
				t.Fatalf("get doc err: %v", err)
			}

			want := map[string]any{
				"user_id": int64(234),
				"a":       "has answer",
				"bet":     tt.wantBet,
				"all_in":  tt.wantAllIn,
			}
			if diff := cmp.Diff(want, d.Data(), cmpopts.EquateApproxTime(time.Second)); diff != "" {
				t.Errorf("bet mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSaveBetWithSpeedData(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	c.Collection("answers").Doc("235:test-q-111").Set(ctx, map[string]any{
		"user_id":  235,
		"a":        "has answer",
		"is_right": true,
	})

	c.Collection("users").Doc("235").Set(ctx, map[string]any{
		"avg_speed": 1.5,
		"speed_cnt": 2,
	})

	f := fstore.New(c)

	err := f.SaveBet(ctx, 235, "test-q-109", 10)
	if err != nil {
		t.Fatalf("fstore save bet: got err %q; want nil", err)
	}

	d, _ := c.Collection("users").Doc("235").Get(ctx)

	if got, want := d.Data()["avg_speed"].(float64), float64(2); got != want {
		t.Errorf("avg speed: got %v; want %v", got, want)
	}
	if got, want := d.Data()["speed_cnt"].(int64), int64(3); got != want {
		t.Errorf("avg speed: got %v; want %v", got, want)
	}
}
