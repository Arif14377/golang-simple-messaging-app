---
title: Bootstrap
keywords: [bootstrap, gorm, validator, env]
description: Integrating Bootstrap.
---

# Bootstrap

[![Github](https://img.shields.io/static/v1?label=&message=Github&color=2ea44f&style=for-the-badge&logo=github)](https://github.com/gofiber/recipes/tree/master/bootstrap) [![StackBlitz](https://img.shields.io/static/v1?label=&message=StackBlitz&color=2ea44f&style=for-the-badge&logo=StackBlitz)](https://stackblitz.com/github/gofiber/recipes/tree/master/bootstrap)

Fiber bootstrap for rapid development using Go-Fiber / Gorm / Validator.

## Components
* Fiber
  * Html Engine Template
  * Logger
  * Monitoring
  * WebSocket (chat)
* Gorm (MySQL)
  * User & session
* MongoDB (`mongo-driver` v2)
  * Riwayat pesan chat
* Validator
* Env File

## Chat & MongoDB

Pesan chat disimpan di MongoDB sebelum di-broadcast ke semua client. Identitas
pengirim (`user_id`, `username`) dan `created_at` diambil dari sesi/JWT di server,
bukan dari payload client.

Collection dan index dibuat otomatis saat aplikasi start:

* database: `MONGODB_DB` (default `simple_messaging_app`)
* collection: `MONGODB_COLLECTION` (default `messages`)
* index: `room_id_1_created_at_-1`

Contoh menjalankan MongoDB dengan auth untuk development:

```bash
docker run -d --name mongo-dev -p 27017:27017 \
  -e MONGO_INITDB_ROOT_USERNAME=root \
  -e MONGO_INITDB_ROOT_PASSWORD=1 \
  -v mongo_data:/data/db \
  mongo:9
```

Lalu isi `.env`:

```dotenv
MONGODB_URI="mongodb://root:1@localhost:27017/?authSource=admin&directConnection=true"
MONGODB_DB="simple_messaging_app"
MONGODB_COLLECTION="messages"
```

Cek data yang tersimpan:

```bash
docker exec mongo-dev mongosh \
  "mongodb://root:1@localhost:27017/simple_messaging_app?authSource=admin" \
  --quiet --eval 'db.messages.find().sort({created_at:-1}).limit(5)'
```

Endpoint riwayat (butuh cookie sesi / Bearer token):

```
GET /message/v1/history?room=lobby&limit=50&before=<RFC3339>
```

WebSocket chat: `GET /message/v1/send` (juga butuh sesi valid). Pesan dikirim
sebagai `{"roomId":"lobby","message":"..."}`; server membalas
`{"id","roomId","from","userId","message","createdAt"}` atau
`{"type":"error","message":"..."}` bila pesan ditolak.

## Router
API Router `/api` with rate limiter middleware
Http Router `/` with CORS and CSRF middleware

## Setup

1. Copy the example env file over:
    ```
    cp .env.example .env
    ```

2. Modify the env file you just copied `.env` with the correct credentials for MySQL **and MongoDB**. Make sure the MySQL database entered in `DB_NAME` has been created; the MongoDB database/collection is created automatically on startup.

3. Run the API:
    ```
    go run main.go
    ```
Your api should be running at `http://localhost:4000/` if the port is in use you may modify it in the `.env` you just created.
