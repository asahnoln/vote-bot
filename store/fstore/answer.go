package fstore

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/asahnoln/vote-bot/store"
)

const AnswersCollection = "answers"

func (f *Firestore) SaveAnswer(ctx context.Context, userID int, questionID string, answer string) error {
	q, err := f.question(ctx, questionID)
	if err != nil {
		return fmt.Errorf("question: %w", err)
	}

	_, err = f.c.Collection("answers").
		Doc(answerDocID(userID, questionID)).
		Set(ctx,
			map[string]any{
				"question_id": q.ID,
				"user_id":     userID,
				"a":           answer,
				"is_right":    q.Opts[answer],
				"answered_at": firestore.ServerTimestamp,
			})
	if err != nil {
		return fmt.Errorf("doc set: %w", err)
	}
	return nil
}

func (f *Firestore) question(ctx context.Context, questionID string) (q store.Question, err error) {
	doc, err := f.c.Collection(QuestionsCollection).Doc(questionID).Get(ctx)
	if err != nil {
		return q, fmt.Errorf("get: %w", err)
	}

	err = doc.DataTo(&q)
	if err != nil {
		return q, fmt.Errorf("data to: %w", err)
	}

	q.ID = doc.Ref.ID

	return q, nil
}

func (f *Firestore) SaveBet(ctx context.Context, userID int, questionID string, bet int) error {
	allIn := false
	if bet == -100 {
		accs, err := f.Accounts(ctx, []int{userID})
		if err != nil {
			return fmt.Errorf("accounts: %w", err)
		}
		bet = accs[userID]
		allIn = true
	}
	_, err := f.c.Collection(AnswersCollection).
		Doc(answerDocID(userID, questionID)).
		Set(ctx,
			map[string]any{
				"bet":    bet,
				"all_in": allIn,
			}, firestore.MergeAll)
	if err != nil {
		return fmt.Errorf("doc set: %w", err)
	}

	q, err := f.queryAndMap[store.Question](ctx, QuestionsCollection, questionID)
	if err != nil {
		return fmt.Errorf("user query: %w", err)
	}

	u, err := f.queryAndMap[store.User](ctx, UsersCollection, strconv.Itoa(userID))
	if err != nil {
		return fmt.Errorf("user query: %w", err)
	}

	a, err := f.queryAndMap[store.Answer](ctx, AnswersCollection, answerDocID(userID, questionID))
	if err != nil {
		return fmt.Errorf("answer query: %w", err)
	}

	if a.IsRight {
		_, err = f.c.Collection(UsersCollection).
			Doc(strconv.Itoa(userID)).
			Set(ctx,
				map[string]any{
					"right_answer_cnt": firestore.Increment(1),
					"avg_speed":        (u.AvgSpeed*float64(u.RightAnswerCnt) + time.Since(q.SentAt).Seconds()) / (float64(u.RightAnswerCnt) + 1),
				}, firestore.MergeAll)
		if err != nil {
			return fmt.Errorf("doc set: %w", err)
		}
	}

	return nil
}

func answerDocID(userID int, questionID string) string {
	return fmt.Sprintf("%d:%s", userID, questionID)
}

func (f *Firestore) queryAndMap[T any](ctx context.Context, col string, id string) (t T, err error) {
	d, err := f.c.Collection(col).
		Doc(id).
		Get(ctx)
	if err != nil {
		return t, fmt.Errorf("get: %w", err)
	}

	err = d.DataTo(&t)
	if err != nil {
		return t, fmt.Errorf("data to: %w", err)
	}

	return t, nil
}
