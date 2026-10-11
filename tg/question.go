package tg

import (
	"fmt"
	"net/http"
)

const OptCallbackDataPrefix = "opt:"

func (b *Bot) question(w http.ResponseWriter, r *http.Request, upd Update) (data any, err error) {
	if b.Store == nil {
		return nil, ErrStoreNil
	}

	q, err := b.Store.CurrentQuestion(r.Context())
	if err != nil {
		return nil, fmt.Errorf("question store: %w", err)
	}

	msg := SendRichMessage{
		ChatID: upd.Message.Chat.ID,
		RichMessage: InputRichMessage{
			Blocks: []InputRichBlock{
				{
					Type: InputRichBlockParagraphType,
					Text: q.Q,
				},
				{
					Type:  InputRichBlockButtonsType,
					Align: InputRichBlockButtonsAlignCenter,
				},
			},
		},
	}
	for _, o := range q.Opts {
		// TODO: TDD utility for button auto row placement
		msg.RichMessage.Blocks[1].Buttons = append(
			msg.RichMessage.Blocks[1].Buttons,
			RichMessageButton{Text: o, CallbackData: OptCallbackDataPrefix + o},
		)
	}

	return msg, nil
}
