[![tests](https://github.com/zaz600/suomen-botti/actions/workflows/go-pr-check.yml/badge.svg)](https://github.com/zaz600/suomen-botti/actions/workflows/go-pr-check.yml)

# Suomen Botti

## Сборка и тестирование
Для работы бота нужен токен от бота, полученный у https://t.me/BotFather

Токен можно передать через переменную окружения `SUOMEN_BOTTI_TG_TOKEN`.
А можно создать файл `.env` и поместить туда строку `SUOMEN_BOTTI_TG_TOKEN=<токен>`

- `make run` - сборка и запуск докер образа с сервером. Запуск осуществляется в фоне.
- `make run-log` - сборка и запуск докер образа с сервером. Не отсоединяется от консоли.
- `make stop` - остановка докер образа
- `make build` - сборка бинарника
- `make lint` - запуск линтера
- `make test` - запуск юнит-тестов

## License

[MIT](http://zaz600.mit-license.org) 
