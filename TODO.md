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
- [x] Admin sends question with answers
- [x] Check for webhook url
- [x] Store for questions
- [x] Ask Alina about bets - are they different for questions?
- [x] Use firebase emulator?
- [x] Receive and save answer and then send bets
- [x] Receive and save bet
- [x] Send question ID with answers
- [x] Remove redundant bot
- [x] Check unhappy cases for answer http and store
- [x] Check unhappy cases for bet http and store
- [x] Send question to those who joined
- [x] Save users who join the bot
- [x] Admin closes question - update latest message to stop accepting answers
- [x] Admin sends answer and account status to all
- [x] Robust messages (what question, answer and bet was chosen at each step and after question closing)
- [x] Can join until first nonoptional question is closed
- [x] Refactor tests
- [x] Fix condition on question handler
- [x] Stop sending questions to those who lost the game
- [x] On reveal we have to define who lost
- [x] If user misses a question - he looses
- [x] Part of errors can be returned to admin without problems
- [x] Bot commands must be scoped to admins
- [ ] VABANK ALL IN
- [ ] Stats MOLNIYA - most fastest right answers (even with 0 account)
- [ ] Stats SHERLOCK - most right answers (even with 0 account)
- [ ] Stats IGROK VA-BANK - most vabank uses
- [ ] Start bot with secret token from QR
- [ ] Delete incoming user messages
- [ ] Think on misclicks on commands - suddenly close a question, or send answer again etc
- [ ] Check default behaviour of bot without handlers (no condition met)
- [ ] Check out whole code and refactor things if necessary
- [ ] Refactor testing destination called
- [ ] Auto register webhook?
- [ ] Optional questions which don't count
- [ ] Check if question was delivered and if the command was repeated - send it if it was not delivered
- [ ] What if EVERYONE LOST before the end of the game??
- [ ] Set webhook with secret token

## Flow questions

- [x] What happens when someone joins midgame?
- [x] What happens when someone runs out of money?
- [x] What happens when someone misses a question? (answer or bet)

## Stats

BOGACH - the richest user
MOLNIYA - most fastest right answers (even with 0 account)
SHERLOCK - most right answers (even with 0 account)
IGROK VA-BANK - most vabank uses
HINTMAN - who use the most hints (even with 0 account)
