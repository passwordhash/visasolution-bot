# Visasolution bot

Это бот, разработанный на языке Go с использованием Selenium, предназначенный для выполнения конкретной задачи: автоматизации процесса записи на визовые собеседования на получение Испанской визы.

<hr/>

![Golang](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Selenium](https://img.shields.io/badge/Selenium-43B02A?style=for-the-badge&logo=selenium&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)

## Требования :clipboard:
- Go версии 1.21 или выше
- Docker, docker-compose для контейнеризации и упрощения развертывания
- Наличие ChatGPT API key и других необходимых данных для работы бота *[(см. конфигурация прокта)](#конфигурация-проекта)*

## Развертывание проекта :hammer_and_wrench:

###  Через Docker Compose 

1. Склонировать репозиторий

    ```bash
    $ git clone https://github.com/passwordhash/visasolution-bot.git
    $ cd visasolution-bot
    ```
   
2. Выполнить необходимую конфигурацию проекта *[(см. конфигурация прокта)](#конфигурация-проекта)*.

    ```bash
    $ cp .env.example .env
    $ cp proxies.json proxies.json.example
    ```

3. Создать и запустить контейнеры :rocket:

    ```bash
    $ docker-compose up -d
    ```
  
## Конфигурация проекта :gear:

### Конфигурация переменных окружения

:exclamation: Создать файл `.env` на основе `.env.example` и заполнить его значениями.

```bash
$ cp .env.example .env
$ vim .env
```

Пояснение к некоторым переменным окружения:

| Переменная           | Описание                                                                                             |
|----------------------|------------------------------------------------------------------------------------------------------|
| `MAIN_LOOP_INTERVAL` | Интервал между итерациями основного цикла бота.                                                      |
| `NOTIFIED_EMAIL`     | Email для отправки уведомлений о результате работы бота.                                             |
| `CHAT_API_KEY`       | API-ключ ChatGPT. Получить можно [здесь](https://platform.openai.com/).                              |
| `SMTP_...`           | Данные для подключения к SMTP-серверу.                                                               |
| `BLS_...`            | Данные для авторизации на сайте BLS.                                                                 |
| `IMGUR_...`          | Секреты для работы с API сервиса [Imgur](https://apidocs.imgur.com/).                                |
| `PROXY_ROW_FOREIGN`  | Прокси для работы с иностранными сайтами (например, ChatGPT). Пример: `host:port@username:password`. |

:exclamation: Также необходимо добавить хотябы 1 российский прокси в файл `proxies.json` на основе `proxies.json.example`.

```bash
$ cp proxies.json.example proxies.json
$ vim proxies.json
```

> **Примечание :bangbang::** Объявление каждой переменной окружения в файле `.env` необходимо для корректной работы бота.

## Работа с логами :card_index_dividers:

Логи сохраняются в директории `/app/logs` внутри контейнера. Эта директория подключена к объявленному в `docker-compose.yml` тому `logs`, что обеспечивает сохранение логов вне контейнера и их доступность даже после перезапуска.

Для просмотра логов бота можно использовать команду `docker-compose logs -f visasolution-bot`.

## Автор :bust_in_silhouette:

студент МГТУ им Н.Э. Баумана ИУ7

Ярослав @prostoYaroslav   

