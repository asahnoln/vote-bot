package tg

import "fmt"

func SpreadButtons(opts map[string]bool, questionID string) (btns [][]RichMessageButton) {
	row := make([]RichMessageButton, 0)
	for o := range opts {
		row = append(row, RichMessageButton{
			Text:         o,
			CallbackData: fmt.Sprintf("%s%s:%s", OptCallbackDataPrefix, questionID, o),
		})
	}
	btns = append(btns, row)
	return
}
