package tg

import "context"

type Wrapper struct {
	Handler Handler
}

func (w *Wrapper) Handle(ctx context.Context, upd Update) (msgs []Request, err error) {
	return w.Handler.Handle(ctx, upd)
}

func (w *Wrapper) Condition(upd Update) bool {
	return w.Handler.Condition(upd)
}

func (w *Wrapper) Name() string {
	return w.Handler.Name()
}

func (w *Wrapper) HandleResponse(ctx context.Context, m any) error {
	return w.Handler.HandleResponse(ctx, m)
}
