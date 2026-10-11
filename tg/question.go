package tg

import (
	"fmt"
	"net/http"
)

func (b *Bot) question(w http.ResponseWriter, r *http.Request, upd Update) {
	if b.Store == nil {
		b.error(ErrStoreNil, w, http.StatusInternalServerError)
		return
	}

	q, err := b.Store.CurrentQuestion(r.Context())
	if err != nil {
		b.error(fmt.Errorf("question store: %w", err), w, http.StatusInternalServerError)
		return
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
			RichMessageButton{Text: o, CallbackData: o},
		)
	}

	b.Dest.Send(r.Context(), msg)
}
