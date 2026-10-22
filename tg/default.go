package tg

import "context"

type DefaultHandler struct{}

func (h *DefaultHandler) Condition(Update) bool {
	return true
}

func (h *DefaultHandler) HandleResponse(context.Context, any) error {
	return nil
}

func (h *DefaultHandler) Name() string {
	return "default"
}

func (h *DefaultHandler) Handle(ctx context.Context, upd Update) (msgs []URLer, err error) {
	msgs = append(msgs, DeleteMessage{
		ChatID:    upd.Message.Chat.ID,
		MessageID: upd.Message.MessageID,
	})
	return msgs, nil
}
