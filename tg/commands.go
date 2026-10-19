package tg

type SetMyCommands struct {
	Commands []BotCommand
	Scope    BotCommandScope
}

type BotCommand struct {
	Command     string
	Description string
}

type BotCommandScopeType string

const ChatScope BotCommandScopeType = "chat"

type BotCommandScope struct {
	Type   BotCommandScopeType
	ChatID int
}

func (SetMyCommands) URL() string {
	return "setMyCommands"
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
