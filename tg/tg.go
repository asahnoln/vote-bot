package tg

import (
	"bytes"
	"encoding/json"
	"net/http"
)

const ApplicationJSONContentType = "application/json"

const (
	SendRichMessageMethod = "sendRichMessage"
)

type SendRichMessage struct {
	ChatID      int
	RichMessage InputRichMessage
}

type InputRichMessage struct {
	Blocks []InputRichBlock
}

type InputRichBlock struct {
	Type    string
	Text    string
	Align   string
	Buttons []RichMessageButton
}

const (
	InputRichBlockParagraphType = "paragraph"
	InputRichBlockButtonsType   = "buttons"
)

const (
	InputRichBlockButtonsAlignCenter = "center"
)

type RichMessageButton struct {
	Text string
}

type Update struct {
	Message Message
}

type Message struct {
	Chat Chat
	Text string
}

type Chat struct {
	ID int
}

type Question struct {
	Q    string
	Opts []string
}

type CurrentQuestioner interface {
	CurrentQuestion() (Question, error)
}

type Bot struct {
	Store CurrentQuestioner
	URL   string
}

func (b *Bot) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	upd := Update{}
	json.NewDecoder(r.Body).Decode(&upd)
	defer r.Body.Close()

	q, _ := b.Store.CurrentQuestion()

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
		msg.RichMessage.Blocks[1].Buttons = append(msg.RichMessage.Blocks[1].Buttons, RichMessageButton{Text: o})
	}

	body, _ := json.Marshal(&msg)

	http.Post(b.URL+"/"+SendRichMessageMethod, ApplicationJSONContentType, bytes.NewReader(body))
}
