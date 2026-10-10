# TODO

I start a bot.
Bot sends me a question with answers.
I choose an answer.
I choose a bet.
At some point, bot sends an answer if I'm right.

In the end, bot can show stats.

        Input
        |
        v
Store -> Game -> Output

## Main

```go
func main() {
  store := db.New()
  b := bot.New(store)
  t := tg.New()
  b.InputFrom(t)
  b.OutputTo(t)

  for err = bot.Serve(); err != nil {
    log.Fatal(err)
  }
}
```

## Flow

User starts Bot.
Bot welcomes user with 1000$.

### Loop

Admin tells Bot to send a question.
Bot sends question to all available users.
User sends their answer and bet to the Bot.
Bot updates user's data - their answer, bet, account.
Admin tells Bot to close the question.
Bot closes the question in DB.
Bot warns all available users that question is closed.
Admin tells Bot to send right answer.
Bot sends users their results - whether their answer is right, how much they won, their current account status.

### End

Admin tells Bot to send stats.
Bot sends users stats.

## Logic

Place a bet - if answer is right, return twice, if it's wrong - return nothing
If user is the fastest to answer - double it.

## Ideas

- [ ] Let a user pass a question? Without bets etc...
- [ ] For your money buy a hint (e.g. either a message from a bot, or remove wrong answer/s etc.)

## Data

Each user has an account. (At the start they have 1000$)

## Telegram

- [x] Auto deploy
- [x] Register webhook
- [ ] Auto register webhook?
- [x] Admin sends question with answers
- [x] Check for webhook url
- [ ] Store for questions
- [ ] Receive and save answer
- [ ] Receive and save bet
- [ ] Save users who join the bot
- [ ] Send question to those who joined (and didn't stop?)
- [ ] Admin close question - update latest message to stop accepting answers
- [ ] What if someone didn't answer?
- [ ] Admin sends answer and account status to all

## Stats

BOGACH - the richest user
MOLNIYA - most fastest right answers (even with 0 account)
SHERLOCK - most right answers (even with 0 account)
IGROK VA-BANK - most vabank uses
HINTMAN - who use the most hints (even with 0 account)
