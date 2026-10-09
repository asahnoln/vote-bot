package tg

import (
	"bytes"
	"encoding/json"
	"net/http"
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
	msg := SendRichMessage{
		ChatID: 123,
		RichMessage: InputRichMessage{
			Blocks: []InputRichBlock{
				{
					Type: "paragraph",
					Text: "X or Y?",
				},
				{
					Type:  "buttons",
					Align: "center",
					Buttons: []RichMessageButton{
						{
							Text: "X",
						},
						{
							Text: "Y",
						},
					},
				},
			},
		},
	}
	body, _ := json.Marshal(&msg)

	http.Post(b.URL+"/sendRichMessage", "application/json", bytes.NewReader(body))
}
