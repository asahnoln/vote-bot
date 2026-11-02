package tg

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/asahnoln/vote-bot/store"
)

var ErrUserIsNotAdmin = errors.New("user is not admin")

type AdminWrapper struct {
	Wrapper
	Store Adminser
}

func (w *AdminWrapper) Handle(ctx context.Context, upd Update) (msgs []Request, err error) {
	admins, err := w.Store.Admins(ctx)
	if err != nil {
		return nil, fmt.Errorf("admins: %w", err)
	}

	if i := slices.IndexFunc(admins, func(u store.User) bool {
		return u.UserID == upd.Message.Chat.ID
	}); i == -1 {
		return nil, fmt.Errorf("user %d: %w", upd.Message.MessageID, ErrUserIsNotAdmin)
	}

	return w.Handler.Handle(ctx, upd)
}
