# Telegram alarm bot
An easy way to manage tasks and reminders in telegram.

## Deploy
Edit `.env.example` file and save it as `.env` file, use dockerfile to run.

## Deploy on [fandogh PaaS](https://www.fandogh.cloud)
Set mongo-uri and bot-token **ENV Secrets**.

## TODO
- [x] Multi-language
- [x] Supports both Date and Duration to add tasks
- [x] Supports different timezones
- [x] Supports Gregorian and Hijri Calenders
- [x] daily reminders
- [x] Task lists
- [x] Automated deploy with github actions
- [ ] Deploy scripts

### Cache service isn't correctly implemented yet, Current Cache interface has memmory leak 

Try bot [here](https://t.me/stillunluckybot).