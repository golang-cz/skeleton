# skeleton

Skeleton app for all our go backends

[Convo example branch](https://github.com/golang-cz/skeleton/tree/convo_example)

- Tools:

  - [go-chi/chi](https://github.com/go-chi/chi)

  - db connector - [upper/db](https://upper.io/v4/)

    - now v3
    - try to use v4

  - migrations - [pressly/goose](https://github.com/pressly/goose)

    - setup goose migrations

  - sentry [sentry-go](https://github.com/getsentry/sentry-go/)

    - [docs](https://pkg.go.dev/github.com/getsentry/sentry-go)

  - toml parser [BurntSushi/toml](https://github.com/BurntSushi/toml)

- Run http server on port, which is defined in etc/config.toml

- Logging

  - Local enviroment: unstructured logging

    - Straight to console?

  - Production enviroment: structured logging (json)

    - Straight to console?

- Implementation of gracefull shutdown of http server

  - sigterm stop server [Go by Example: Signals](https://gobyexample.com/signals)

- Database:

  - PostgreSQL?
  - Hosted in Docker with docker-compose

- Configuration https middlewares:

  - RealIP
  - RequestLogger
  - Recoverer
    - Local: run it
    - Production: log it to sentry and kibana

- Status page

  - Status info
    - db
    - information about go application
  - ![Example](status_page_example.png)
