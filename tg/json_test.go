package tg_test

import (
	_ "embed"
	"encoding/json"
	"strings"
	"testing"

	"github.com/asahnoln/vote-bot/tg"
	"github.com/google/go-cmp/cmp"
)

//go:embed testdata/sendRichMessage.json
var sendRichMessage string

func TestSendRichMessageStr(t *testing.T) {
	msg := tg.SendRichMessage{
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
	got, err := json.MarshalIndent(&msg, "", "  ")
	if err != nil {
		t.Fatalf("got err %v; want nil", err)
	}

	want := strings.TrimSpace(sendRichMessage)
	if diff := cmp.Diff(want, string(got)); diff != "" {
		t.Errorf("send rich message mismatch (-want +got):\n%s", diff)
	}
}
