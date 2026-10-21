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
							Text:         "button 1",
							CallbackData: "button data",
						},
					},
					Cells: [][]tg.RichBlockTableCell{
						{
							{
								Text:  "table cell 1",
								Align: "left",
							},
						},
					},
					Caption:   "caption",
					IsStriped: true,
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

//go:embed testdata/editMessageText.json
var editMessageText string

func TestEditMessageTextJSON(t *testing.T) {
	msg := tg.EditMessageText{
		ChatID:    12,
		MessageID: 13,
		RichMessage: tg.InputRichMessage{
			Blocks: []tg.InputRichBlock{
				{
					Type:  "editType",
					Text:  "edit text",
					Align: "right",
					Buttons: []tg.RichMessageButton{
						{
							Text:         "edit button 1",
							CallbackData: "edit button data",
						},
					},
					Cells: [][]tg.RichBlockTableCell{
						{
							{
								Text:  "edit table cell 1",
								Align: "center",
							},
						},
					},
					Caption:   "edit caption",
					IsStriped: true,
				},
			},
		},
	}
	got, err := json.MarshalIndent(&msg, "", "  ")
	if err != nil {
		t.Fatalf("got err %v; want nil", err)
	}

	want := strings.TrimSpace(editMessageText)
	if diff := cmp.Diff(want, string(got)); diff != "" {
		t.Errorf("send rich message mismatch (-want +got):\n%s", diff)
	}
}

//go:embed testdata/setMyCommands.json
var setMyCommands string

func TestSetMyCommandsJSON(t *testing.T) {
	msg := tg.SetMyCommands{
		Commands: []tg.BotCommand{
			{
				Command:     "/start",
				Description: "Начать",
			},
		},
		Scope: tg.BotCommandScope{
			Type:   tg.ChatScope,
			ChatID: 200,
		},
	}
	got, err := json.MarshalIndent(&msg, "", "  ")
	if err != nil {
		t.Fatalf("got err %v; want nil", err)
	}

	want := strings.TrimSpace(setMyCommands)
	if diff := cmp.Diff(want, string(got)); diff != "" {
		t.Errorf("set my commands mismatch (-want +got):\n%s", diff)
	}
}

//go:embed testdata/deleteMessage.json
var deleteMessage string

func TestDeleteMessageJSON(t *testing.T) {
	msg := tg.DeleteMessage{
		ChatID:    299,
		MessageID: 12,
	}
	got, err := json.MarshalIndent(&msg, "", "  ")
	if err != nil {
		t.Fatalf("got err %v; want nil", err)
	}

	want := strings.TrimSpace(deleteMessage)
	if diff := cmp.Diff(want, string(got)); diff != "" {
		t.Errorf("set my commands mismatch (-want +got):\n%s", diff)
	}
}
