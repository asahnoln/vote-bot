package tg

type BotCommand struct {
	Command     string `json:"command"`
	Description string `json:"description"`
}

type BotCommandScopeType string

const ChatScope BotCommandScopeType = "chat"

type BotCommandScope struct {
	Type   BotCommandScopeType `json:"type"`
	ChatID int                 `json:"chat_id"`
}

type SetMyCommands struct {
	Commands []BotCommand    `json:"commands"`
	Scope    BotCommandScope `json:"scope"`
}

func (SetMyCommands) URL() string {
	return "setMyCommands"
}

func (SetMyCommands) UnmarshalResult(data []byte) (any, error) {
	return nil, nil
}

const (
	SendCurrentQuestionCommand  = "/sendcurrentquestion"
	CloseCurrentQuestionCommand = "/closecurrentquestion"
	SendAnswerCommand           = "/sendanswer"
	StatsRichestUsersCommand    = "/statsrichestusers"
)

func CommandsForAdmin(id int) SetMyCommands {
	return SetMyCommands{
		Commands: []BotCommand{
			{
				Command:     SendCurrentQuestionCommand,
				Description: "Выслать текущий вопрос",
			},
			{
				Command:     CloseCurrentQuestionCommand,
				Description: "Закрыть текущий вопрос",
			},
			{
				Command:     SendAnswerCommand,
				Description: "Выслать всем ответ",
			},
			{
				Command:     StatsRichestUsersCommand,
				Description: "Статистика: богачи",
			},
		},
		Scope: BotCommandScope{
			Type:   ChatScope,
			ChatID: id,
		},
	}
}
