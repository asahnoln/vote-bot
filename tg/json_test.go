package tg_test

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/asahnoln/vote-bot/tg"
	"github.com/google/go-cmp/cmp"
)

//go:embed testdata/sendRichMessage.json
var sendRichMessage []byte

func TestSendRichMessage(t *testing.T) {
	want := tg.SendRichMessage{
		ChatID: 67,
		RichMessage: tg.InputRichMessage{
			Blocks: []tg.InputRichBlock{
				{
					Type:  "someType",
					Text:  "some text",
					Align: "left",
					Buttons: []tg.RichMessageButton{
						{
							Text: "button 1",
						},
					},
				},
			},
		},
	}
	got := tg.SendRichMessage{}
	err := json.Unmarshal(sendRichMessage, &got)
	if err != nil {
		t.Fatalf("got err %v; want nil", err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("send rich message mismatch (-want +got):\n%s", diff)
	}
}
