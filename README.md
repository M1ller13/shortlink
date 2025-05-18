 ShortLink — сервис сокращения URL

[![Go](https://img.shields.io/badge/Go-1.21+-blue?logo=go)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-24.0+-blue?logo=docker)](https://docker.com)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)

Микросервис для сокращения ссылок с аналитикой переходов. Аналог Bitly, написанный на Go с использованием Gin, PostgreSQL и Redis.

 🚀 Возможности
- Сокращение длинных URL → генерация коротких ссылок.
- Редирект по коротким ссылкам.
- Аналитика переходов (количество кликов, IP, дата).
- Кэширование через Redis для ускорения ответов.
- Docker-контейнеризация (легкий деплой).

 🛠 Стек технологий
- Язык: Go 1.21+
- Фреймворк: Gin
- Базы данных: PostgreSQL, Redis
- Контейнеризация: Docker, Docker Compose
- Дополнительно: JWT, Swagger (опционально)

 📦 Установка и запуск

 1. Клонируйте репозиторий
```bash
git clone https://github.com/M1ler13/shortlink.git
cd shortlink
