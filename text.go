package main

import (
	"strconv"

	"github.com/asahnoln/vote-bot/store"
)

type textMap map[string]string

func (t textMap) TextByKey(key string) string {
	return t[key]
}

var ruText textMap = textMap{
	"startBalanceAgain %d": "Ваш баланс снова %d$. Началась реальная игра!",
	"welcomeToTheGame":     "💸 Ваш стартовый капитал " + strconv.Itoa(store.AccountStartMoney) + "$\nСтавьте с умом, играйте смело — и пусть удача будет на вашей стороне!",
	"missedFirstQuestion":  "Вы пропустили первый вопрос. Правила есть правила: в этот раз вы не участвуете.",
	"alreadyJoinedTheGame": "Вы уже в игре!\nПовторный вход не требуется.",
}
