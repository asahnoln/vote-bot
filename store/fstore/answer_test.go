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

func TestTutorialSaveAnswer(t *testing.T) {
	ctx := context.Background()
	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	c.Collection("config").Doc("mode").Set(ctx, map[string]any{"mode": store.TutorialMode})
	c.Collection("tutorial_questions").Doc("tut-q-67").Set(ctx, store.Question{
		Q:     "52 or 42?",
		Opts:  map[string]bool{"wrong": false, "right": true},
		Order: 0,
	})

	f := fstore.New(c)

	err := f.SaveAnswer(ctx, 51, "tut-q-67", "right")
	if err != nil {
		t.Fatalf("fstore save tutorial answer: got err %q; want nil", err)
	}

	d, err := c.Collection("tutorial_answers").Doc("51:tut-q-67").Get(ctx)
	if err != nil {
		t.Fatalf("get doc err: %v", err)
	}

	if got, want := d.Data()["user_id"], int64(51); got != want {
		t.Errorf("user id: got %v; want %v", got, want)
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

	c.Collection("config").Doc("current_question").Set(ctx, map[string]any{
		"idx": 0,
	})

	c.Collection("questions").Doc("test-q-111").Set(ctx, map[string]any{})

	for _, tt := range tests {
		t.Run(fmt.Sprintf("bet %d", tt.bet), func(t *testing.T) {
			c.Collection("users").Doc("234").Set(ctx, map[string]any{"user_id": 234, "acc": 1000})
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

func TestTutorialSaveBet(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	c.Collection("config").Doc("current_question").Set(ctx, map[string]any{
		"idx": 0,
	})
	c.Collection("config").Doc("mode").Set(ctx, map[string]any{
		"mode": store.TutorialMode,
	})

	c.Collection("tutorial_questions").Doc("tut-q-111").Set(ctx, map[string]any{})
	c.Collection("users").Doc("233").Set(ctx, map[string]any{"user_id": 233, "acc": 1000})
	c.Collection("tutorial_answers").Doc("233:tut-q-111").Set(ctx, map[string]any{
		"user_id": 233,
		"a":       "has tutorial answer",
	})

	f := fstore.New(c)

	err := f.SaveBet(ctx, 233, "tut-q-111", 10)
	if err != nil {
		t.Fatalf("fstore save bet: got err %q; want nil", err)
	}

	d, err := c.Collection("tutorial_answers").Doc("233:tut-q-111").Get(ctx)
	if err != nil {
		t.Fatalf("get doc err: %v", err)
	}

	if got, want := d.Data()["bet"], int64(10); got != want {
		t.Errorf("user id: got %v; want %v", got, want)
	}
}

func TestSaveBetWithStatsData(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	c.Collection("questions").Doc("test-q-109").Set(ctx, map[string]any{
		"sent_at": time.Now().Add(-3 * time.Second),
	})

	c.Collection("answers").Doc("235:test-q-109").Set(ctx, map[string]any{
		"user_id":  235,
		"is_right": true,
	})

	c.Collection("users").Doc("235").Set(ctx, map[string]any{
		"right_answer_cnt": 2,
		"avg_speed":        1.5,
	})

	f := fstore.New(c)

	err := f.SaveBet(ctx, 235, "test-q-109", 10)
	if err != nil {
		t.Fatalf("fstore save bet: got err %q; want nil", err)
	}

	d, _ := c.Collection("users").Doc("235").Get(ctx)

	if diff := cmp.Diff(d.Data()["avg_speed"], float64(2), cmpopts.EquateApprox(0.01, 0)); diff != "" {
		t.Errorf("avg speed mismatch (-want +got):\n%s", diff)
	}
	if got, want := d.Data()["right_answer_cnt"].(int64), int64(3); got != want {
		t.Errorf("avg speed: got %v; want %v", got, want)
	}
}

func TestSaveBetWrongAnswerWontSaveStats(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	c.Collection("questions").Doc("test-q-108").Set(ctx, map[string]any{
		"sent_at": time.Now().Add(-30 * time.Second),
	})

	c.Collection("answers").Doc("236:test-q-108").Set(ctx, map[string]any{
		"user_id":  236,
		"is_right": false,
	})

	c.Collection("users").Doc("236").Set(ctx, map[string]any{
		"right_answer_cnt": 5,
		"avg_speed":        3,
	})

	f := fstore.New(c)

	err := f.SaveBet(ctx, 236, "test-q-108", 11)
	if err != nil {
		t.Fatalf("fstore save bet: got err %q; want nil", err)
	}

	d, _ := c.Collection("users").Doc("236").Get(ctx)

	if diff := cmp.Diff(d.Data()["avg_speed"], int64(3), cmpopts.EquateApprox(0.001, 0)); diff != "" {
		t.Errorf("avg speed mismatch (-want +got):\n%s", diff)
	}
	if got, want := d.Data()["right_answer_cnt"].(int64), int64(5); got != want {
		t.Errorf("avg speed: got %v; want %v", got, want)
	}
}

func TestSaveBetSavesAcc(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	c.Collection("questions").Doc("test-q-107").Set(ctx, map[string]any{})

	tests := []struct {
		isRight bool
		wantAcc int64
	}{
		{true, 2300},
		{false, 1700},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("right answer %v", tt.isRight), func(t *testing.T) {
			c.Collection("answers").Doc("237:test-q-107").Set(ctx, map[string]any{
				"user_id":  237,
				"is_right": tt.isRight,
			})

			c.Collection("users").Doc("237").Set(ctx, map[string]any{
				"acc": 2000,
			})

			f := fstore.New(c)

			err := f.SaveBet(ctx, 237, "test-q-107", 300)
			if err != nil {
				t.Fatalf("fstore save bet: got err %q; want nil", err)
			}

			d, _ := c.Collection("users").Doc("237").Get(ctx)

			if got, want := d.Data()["acc"].(int64), int64(tt.wantAcc); got != want {
				t.Errorf("avg speed: got %v; want %v", got, want)
			}
		})
	}
}
