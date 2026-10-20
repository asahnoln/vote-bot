package tg

import (
	"fmt"
)

func SpreadButtons(opts map[string]bool, questionID string, maxPerRow int) (btns [][]RichMessageButton) {
	n := len(opts)
	rows := (n + maxPerRow - 1) / maxPerRow
	basePerRow := n / rows
	extraPerRow := n % rows

	extraAdder := 0
	if extraPerRow > 0 {
		extraAdder = 1
	}

	btns = make([][]RichMessageButton, 0, rows)
	row := make([]RichMessageButton, 0, basePerRow+extraAdder)
	for o := range opts {
		if len(row) == cap(row) {
			if extraPerRow > 0 {
				extraPerRow -= 1
			}
			if extraPerRow == 0 {
				extraAdder = 0
			}

			btns = append(btns, row)
			row = make([]RichMessageButton, 0, basePerRow+extraAdder)
		}

		row = append(row, RichMessageButton{
			Text:         o,
			CallbackData: fmt.Sprintf("%s%s:%s", OptCallbackDataPrefix, questionID, o),
		})

	}

	btns = append(btns, row)

	return
}
