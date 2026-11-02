# Simple Messaging App

Aplikasi chat real-time sederhana dengan **Go + Fiber v3**. Autentikasi memakai JWT
dan cookie HttpOnly, data user/sesi disimpan di **MySQL**, riwayat pesan di
**MongoDB**, dan observabilitas memakai **Elastic APM** + **ELK Stack**.

![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)
![Fiber](https://img.shields.io/badge/Fiber-v3-00ACD7)
![MySQL](https://img.shields.io/badge/MySQL-8+-4479A1?logo=mysql&logoColor=white)
![MongoDB](https://img.shields.io/badge/MongoDB-7+-47A248?logo=mongodb&logoColor=white)
![Elastic](https://img.shields.io/badge/Elastic-APM%20%2B%20ELK-005571?logo=elastic&logoColor=white)

---

## Daftar Isi

- [Fitur](#fitur)
- [Arsitektur](#arsitektur)
- [Teknologi](#teknologi)
- [Struktur Proyek](#struktur-proyek)
- [Prasyarat](#prasyarat)
- [Menjalankan Aplikasi](#menjalankan-aplikasi)
- [Konfigurasi](#konfigurasi)
- [API](#api)
- [WebSocket](#websocket)
- [Penyimpanan Data](#penyimpanan-data)
- [Observability](#observability)
- [Perintah Development](#perintah-development)
- [Batasan](#batasan)
- [Lisensi](#lisensi)

---

## Fitur

- **Autentikasi**
  - Register, login, logout, dan profil user.
  - Password di-hash dengan **bcrypt**.
  - **Access token JWT** (15 menit) + **refresh token** (7 hari) dengan rotasi.
  - Token disimpan di **cookie HttpOnly** (`access_token`, `refresh_token`), sehingga
    tidak bisa dibaca JavaScript. Fallback `Authorization: Bearer <token>` tersedia
    untuk client non-browser.
  - Session disimpan di MySQL; logout menghapus session, dan rotasi token mengganti
    hash token yang tersimpan.
- **Chat real-time**
  - WebSocket di `/message/v1/send`, satu room global (`lobby`).
  - Identitas pengirim, timestamp, dan ID pesan **selalu diisi server**, bukan dari client.
  - Pesan disimpan ke MongoDB **sebelum** di-broadcast ke semua client.
  - Riwayat pesan dengan pagination (`limit`, `before`).
  - Frontend melakukan auto-reconnect, refresh token otomatis, dan desktop notification.
- **Operasional**
  - Graceful shutdown (HTTP server + koneksi MongoDB + flush APM).
  - Panic recovery, access log ke file & stdout, rate limiter pada `/api`.
  - CORS + CSRF middleware pada route UI (`/`).
  - **Elastic APM** untuk transaksi dan span per proses.
  - **ELK Stack** (Filebeat → Logstash → Elasticsearch → Kibana) untuk log.
  - Dashboard monitor Fiber di `/dashboard`.

---

## Arsitektur

```text
                        HTTP (cookie HttpOnly)
   Browser  ────────────────────────────────────┐
      │                                          ▼
      │                               ┌──────────────────────┐
      │                               │ Fiber v3             │
      │                               │  APM → recover →     │
      │                               │  logger → auth       │
      │                               │  controller / ws hub │
      │                               └───────┬──────────────┘
      │                                       │
      │            WebSocket                  ├──────────────► MySQL (GORM)
      └──────────────────────────────────────►│                users, user_sessions
                                              │
                                              ├──────────────► MongoDB
                                              │                messages
                                              │
                                              └──────────────► APM Server :8200

   logs/*.log ──► Filebeat ──► Logstash ──► Elasticsearch ──► Kibana
```

**Alur utama**

1. **Register** → validasi input → hash password (bcrypt) → simpan user di MySQL.
2. **Login** → verifikasi password → buat access token JWT + refresh token →
   simpan hash token di `user_sessions` → set cookie HttpOnly.
3. **Request terproteksi** → middleware `Auth` memverifikasi JWT **dan** memastikan
   session masih ada di MySQL (token yang sudah logout tidak bisa dipakai).
4. **Kirim pesan** → WebSocket menerima payload → validasi → simpan ke MongoDB →
   broadcast ke semua client yang terhubung.
5. **Riwayat** → `GET /message/v1/history` membaca MongoDB, lalu mengembalikan pesan
   dalam urutan lama → baru.

---

## Teknologi

| Komponen | Teknologi |
|---|---|
| Bahasa | Go 1.26 |
| Web framework | Fiber v3 |
| Template UI | Fiber HTML template (`views/index.html`) |
| Database relasional | MySQL + GORM (AutoMigrate) |
| Database dokumen | MongoDB (`mongo-driver` v2) |
| Validasi | `go-playground/validator/v10` |
| Auth | `golang-jwt/jwt/v5`, bcrypt |
| Observability | `go.elastic.co/apm/v2`, Elastic APM Server |
| Log pipeline | Filebeat, Logstash, Elasticsearch, Kibana (Docker Compose) |

---

## Struktur Proyek

```text
.
├── main.go                     # entrypoint + graceful shutdown
├── bootstrap/
│   └── bootstrap.go            # setup env, log, DB, middleware, router
├── app/
│   ├── controllers/            # handler HTTP
│   ├── middleware/             # auth + APM transaction middleware
│   ├── models/                 # model GORM & dokumen MongoDB
│   ├── repository/             # query MySQL/MongoDB + span DB
│   └── ws/                     # hub, handler, dan broadcast WebSocket
├── pkg/
│   ├── cookie/                 # helper cookie HttpOnly
│   ├── database/               # koneksi & migrasi MySQL/MongoDB
│   ├── dto/                    # request/response DTO
│   ├── env/                    # loader .env
│   ├── jwt/                    # generate & verify JWT
│   ├── response/               # format response envelope
│   ├── router/                 # registrasi route
│   ├── token/                  # hash token & refresh token
│   └── tracing/                # tracer & helper span Elastic APM
├── views/
│   └── index.html              # halaman login/register + chat
├── elk_stack/                  # Docker Compose & config ELK + APM Server
├── logs/                       # log runtime (di-ignore Git)
├── .env.example
├── go.mod
└── go.sum
```

> Jalankan aplikasi dari **root proyek**, karena `.env`, `./views`, dan `./logs`
> dibaca dengan path relatif.

---

## Prasyarat

- **Go 1.26+**
- **MySQL 8+** — database harus sudah dibuat; tabel dibuat otomatis saat startup.
- **MongoDB 7+** — database dan collection dibuat otomatis saat startup.
- **Docker + Docker Compose** — opsional, untuk menjalankan ELK Stack.
- `curl` dan (opsional) `websocat` untuk mencoba API/WebSocket.

---

## Menjalankan Aplikasi

### 1. Clone repository

```bash
git clone https://github.com/Arif14377/golang-simple-messaging-app.git
cd golang-simple-messaging-app
```

### 2. Siapkan database

Buat database MySQL (nama harus sama dengan `DB_NAME`):

```sql
CREATE DATABASE simple_messaging_app
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;
```

Contoh cepat dengan Docker:

```bash
# MySQL
docker run -d --name mysql-dev -p 3306:3306 \
  -e MYSQL_ROOT_PASSWORD=1 \
  -e MYSQL_DATABASE=simple_messaging_app \
  -e MYSQL_USER=app_user \
  -e MYSQL_PASSWORD=1 \
  mysql:lts

# MongoDB
docker run -d --name mongo-dev -p 27017:27017 \
  -e MONGO_INITDB_ROOT_USERNAME=root \
  -e MONGO_INITDB_ROOT_PASSWORD=1 \
  -v mongo_data:/data/db \
  mongo:9
```

### 3. Buat file `.env`

```bash
cp .env.example .env
```

Sesuaikan nilainya (lihat [Konfigurasi](#konfigurasi)). Contoh untuk Docker di atas:

```dotenv
DB_USER="app_user"
DB_PASSWORD="1"
DB_HOST="127.0.0.1"
DB_PORT="3306"
DB_NAME="simple_messaging_app"

APP_HOST="localhost"
APP_PORT="8888"

JWT_SECRET="ganti-dengan-secret-yang-kuat"
COOKIE_SECURE="false"

MONGODB_URI="mongodb://root:1@localhost:27017/?authSource=admin&directConnection=true"
MONGODB_DB="simple_messaging_app"
MONGODB_COLLECTION="messages"

ELASTIC_APM_SERVICE_NAME="simple-messaging-app"
ELASTIC_APM_SERVER_URL="http://localhost:8200"
ELASTIC_APM_ENVIRONMENT="development"
```

### 4. Siapkan direktori log

Direktori `logs/` tidak ikut Git, jadi buat dulu:

```bash
mkdir -p logs
```

### 5. Jalankan

```bash
go mod download
go run .
```

Aplikasi tersedia di:

| URL | Keterangan |
|---|---|
| <http://localhost:8888/> | UI login/register + chat |
| <http://localhost:8888/dashboard> | Dashboard monitor Fiber |
| <http://localhost:8888/api/> | Contoh endpoint API (rate limited) |
| `ws://localhost:8888/message/v1/send` | WebSocket chat |

---

## Konfigurasi

Semua konfigurasi dibaca dari `.env` saat startup.

| Variabel | Default | Wajib | Keterangan |
|---|---|---|---|
| `DB_USER` | — | Ya | User MySQL |
| `DB_PASSWORD` | — | Ya | Password MySQL |
| `DB_HOST` | `127.0.0.1` | Tidak | Host MySQL |
| `DB_PORT` | `3306` | Tidak | Port MySQL |
| `DB_NAME` | — | Ya | Nama database (harus sudah ada) |
| `APP_HOST` | `localhost` | Tidak | Bind address aplikasi |
| `APP_PORT` | `8888` | Tidak | Port aplikasi |
| `JWT_SECRET` | `secret` | **Ganti** | Secret penandatangan JWT |
| `COOKIE_SECURE` | `false` | Tidak | Set `true` jika berjalan di HTTPS |
| `COOKIE_DOMAIN` | *(kosong)* | Tidak | Domain cookie (opsional) |
| `MONGODB_URI` | `mongodb://localhost:27017` | Ya | Connection string MongoDB |
| `MONGODB_DB` | `simple_messaging_app` | Tidak | Nama database MongoDB |
| `MONGODB_COLLECTION` | `messages` | Tidak | Nama collection pesan |
| `ELASTIC_APM_SERVICE_NAME` | `simple-messaging-app` | Tidak | Nama service di APM |
| `ELASTIC_APM_SERVER_URL` | `http://localhost:8200` | Tidak | URL APM Server |
| `ELASTIC_APM_ENVIRONMENT` | `development` | Tidak | Environment di APM |

Catatan:

- Ganti `JWT_SECRET` di production. Nilai default `secret` hanya untuk development.
- `COOKIE_SECURE=true` wajib jika aplikasi diakses lewat HTTPS.
- Jika APM Server tidak berjalan, aplikasi tetap jalan; event APM hanya tidak terkirim.

---

## API

### Autentikasi

- Endpoint bertanda **Auth** membutuhkan cookie `access_token` **atau** header
  `Authorization: Bearer <access_token>`.
- Access token berlaku **15 menit**; refresh token **7 hari**.
- Response login juga mengembalikan `token` dan `refreshToken` di body, sehingga
  client non-browser bisa memakai header `Authorization: Bearer`.
- `POST /user/v1/refresh-token` merotasi kedua token. Refresh token dibaca dari cookie,
  atau dari body `{"refreshToken":"..."}` untuk client non-browser.
- Logout menghapus session di MySQL dan cookie di browser.

### Format response

Sukses:

```json
{
  "success": true,
  "message": "User profile",
  "data": { "userId": 1, "username": "johndoe", "fullName": "John Doe" }
}
```

Gagal:

```json
{
  "success": false,
  "message": "Invalid or expired token."
}
```

### Endpoint

| Method | Path | Auth | Keterangan |
|---|---|---|---|
| `POST` | `/user/v1/register` | Tidak | Registrasi user baru |
| `POST` | `/user/v1/login` | Tidak | Login, set cookie token |
| `DELETE` | `/user/v1/logout` | Ya | Hapus session + cookie |
| `POST` | `/user/v1/refresh-token` | Refresh token | Rotasi access & refresh token |
| `GET` | `/user/v1/me` | Ya | Profil user yang sedang login |
| `GET` | `/message/v1/history` | Ya | Riwayat pesan per room |
| `GET` | `/message/v1/send` | Ya | Upgrade ke WebSocket chat |
| `GET` | `/` | Tidak | Halaman UI |
| `GET` | `/dashboard` | Tidak | Monitor Fiber |
| `GET` | `/api/` | Tidak | Contoh endpoint (rate limited) |

### Contoh

**Register**

```bash
curl -X POST http://localhost:8888/user/v1/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"johndoe","fullName":"John Doe","password":"secret123"}'
```

**Login** (simpan cookie ke `cookies.txt`)

```bash
curl -c cookies.txt -X POST http://localhost:8888/user/v1/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"johndoe","password":"secret123"}'
```

**Profil**

```bash
curl -b cookies.txt http://localhost:8888/user/v1/me
```

**Riwayat pesan**

```bash
curl -b cookies.txt \
  'http://localhost:8888/message/v1/history?room=lobby&limit=50&before=2026-11-02T12:00:00Z'
```

- `room` default `lobby`.
- `limit` default `50`, maksimal `100`.
- `before` (RFC3339) dipakai untuk pagination ke pesan yang lebih lama.

**Refresh token**

```bash
curl -b cookies.txt -c cookies.txt -X POST \
  http://localhost:8888/user/v1/refresh-token \
  -H 'Content-Type: application/json' -d '{}'
```

**Logout**

```bash
curl -b cookies.txt -X DELETE http://localhost:8888/user/v1/logout
```

---

## WebSocket

- URL: `ws://localhost:8888/message/v1/send` (`wss://` jika memakai HTTPS).
- Handshake memakai cookie sesi yang sama dengan HTTP, jadi client harus terhubung
  dari origin yang sama (atau mengirim cookie secara eksplisit).
- Hanya user dengan sesi valid yang bisa terhubung.

**Client → server**

```json
{ "roomId": "lobby", "message": "Halo semua!" }
```

**Server → semua client**

```json
{
  "id": "673a1f2c9d3e4b5a6c7d8e9f",
  "roomId": "lobby",
  "from": "johndoe",
  "userId": 1,
  "message": "Halo semua!",
  "createdAt": "2026-11-02T12:34:56Z"
}
```

**Pesan ditolak (hanya ke pengirim)**

```json
{ "type": "error", "message": "Pesan tidak boleh kosong." }
```

Aturan:

- `roomId` kosong → otomatis `lobby`.
- Pesan maksimal **2000 karakter** (setelah `TrimSpace`).
- Identitas, timestamp, dan ID diisi server; client tidak bisa memalsukannya.
- Pesan disimpan ke MongoDB dulu; jika gagal, pesan tidak di-broadcast.
- Batas baca frame 16 KiB dan timeout tulis 10 detik.

Contoh dengan `websocat`:

```bash
websocat -H="Cookie: access_token=<token>" \
  ws://localhost:8888/message/v1/send
```

---

## Penyimpanan Data

### MySQL (GORM, AutoMigrate saat startup)

| Tabel | Isi |
|---|---|
| `users` | `id`, `username` (unik), `full_name`, `password` (bcrypt), timestamps |
| `user_sessions` | `id`, `user_id`, `token_hash`, `refresh_token` (hash), `refresh_token_expired`, timestamps |

Token tidak disimpan dalam bentuk asli: access token dan refresh token di-hash
SHA-256 sebelum disimpan.

### MongoDB

| Item | Nilai |
|---|---|
| Database | `MONGODB_DB` (default `simple_messaging_app`) |
| Collection | `MONGODB_COLLECTION` (default `messages`) |
| Index | `room_id: 1`, `created_at: -1` (nama: `room_id_created_at`) |

Dokumen pesan:

```json
{
  "_id": "673a1f2c9d3e4b5a6c7d8e9f",
  "room_id": "lobby",
  "user_id": 1,
  "username": "johndoe",
  "content": "Halo semua!",
  "created_at": "2026-11-02T12:34:56Z"
}
```

Collection dan index dibuat otomatis saat aplikasi startup.

---

## Observability

### Elastic APM

Instrumentasi ada di `pkg/tracing` dan `app/middleware/apm_middleware.go`.

- **Transaction HTTP**: `GET /`, `POST /user/v1/login`, `GET /user/v1/me`, dst.
- **Transaction WebSocket**: `ws.message` (satu per pesan) dan `ws.broadcast`.
- **Span**: controller, middleware, `db` (mysql/mongodb), `crypto` (bcrypt),
  `jwt`, `token`, `validate`, `cookie`, `response`, `websocket`.

Aktifkan lewat `.env`:

```dotenv
ELASTIC_APM_SERVICE_NAME="simple-messaging-app"
ELASTIC_APM_SERVER_URL="http://localhost:8200"
ELASTIC_APM_ENVIRONMENT="development"
```

Data APM bisa dilihat di **Kibana → Observability → APM**, atau di Discover dengan
data view `traces-apm-*`.

### Log

- File: `./logs/simple_messaging_app.log` (juga ditulis ke stdout).
- Fiber access log: method, path, status, latency, IP.
- Log aplikasi: startup, error database, pesan ditolak, panic + stack trace.

### ELK Stack (opsional, development)

```bash
docker compose -f elk_stack/docker-compose.yml up -d
docker compose -f elk_stack/docker-compose.yml ps
```

| Service | URL / Port |
|---|---|
| Elasticsearch | <http://127.0.0.1:9200> |
| Kibana | <http://127.0.0.1:5601> |
| APM Server | <http://127.0.0.1:8200> |
| Logstash | `5044` (internal Docker network) |

Data view Kibana:

- **Log aplikasi**: `simple-messaging-*`
- **APM traces**: `traces-apm-*`

Variabel opsional:

```bash
ELK_VERSION=8.15.3 LOG_TIMEZONE=Asia/Jakarta \
  docker compose -f elk_stack/docker-compose.yml up -d
```

> Stack ini khusus development: security Elasticsearch dimatikan dan semua port
> di-bind ke `127.0.0.1`. Jangan dipakai apa adanya di production.

---

## Perintah Development

```bash
# format
go fmt ./...

# build & static check
go build ./...
go vet ./...
go test ./...

# jalankan
go run .
```

Saat ini belum ada file unit test (`go test ./...` akan melaporkan `no test files`).

---

## Batasan

- Hanya satu room global (`lobby`); UI belum menyediakan multi-room.
- Hub WebSocket disimpan di memori proses. Untuk multi-instance perlu pub/sub
  (mis. Redis) agar broadcast sampai ke semua instance.
- Refresh token disimpan sebagai hash di MySQL; belum ada mekanisme revoke-all-devices.
- ELK Stack hanya untuk development (tanpa TLS/autentikasi).
- Belum ada automated test dan belum ada rate limit khusus untuk endpoint login.

---

## Lisensi

Belum ada file lisensi di repository ini.
