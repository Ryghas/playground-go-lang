# pokemon-service

Boilerplate REST API CRUD (contoh: Pokémon) dengan **Go + Gin + GORM + PostgreSQL**, mengikuti Clean Architecture. Dilengkapi Swagger UI, health check liveness/readiness, migration, seed data, Docker Compose, unit & integration test, serta CI GitHub Actions.

## Daftar Isi

- [Tech Stack](#tech-stack)
- [Prasyarat](#prasyarat)
- [Quick Start](#quick-start)
- [Konfigurasi](#konfigurasi)
- [Perintah Makefile](#perintah-makefile)
- [Dokumentasi API (Swagger)](#dokumentasi-api-swagger)
- [Endpoint](#endpoint)
- [Contoh Request (curl / Postman)](#contoh-request-curl--postman)
- [Format Response & Error](#format-response--error)
- [Aturan Bisnis](#aturan-bisnis)
- [Arsitektur](#arsitektur)
- [Testing](#testing)
- [CI (GitHub Actions)](#ci-github-actions)
- [Troubleshooting](#troubleshooting)
- [Langkah Lanjut](#langkah-lanjut)

## Tech Stack

| Komponen | Library / Tool |
|---|---|
| Bahasa | Go 1.25+ |
| HTTP framework | [Gin](https://github.com/gin-gonic/gin) |
| ORM | [GORM](https://gorm.io) + driver Postgres (pgx) |
| Database | PostgreSQL 16 |
| Migration | [golang-migrate](https://github.com/golang-migrate/migrate) (via Docker) |
| Dokumentasi API | [swaggo/swag](https://github.com/swaggo/swag) + gin-swagger |
| Logging | `log/slog` (JSON) |
| Container | Docker multi-stage, runtime distroless (non-root) |

## Prasyarat

- **Docker Desktop**, dan pastikan statusnya *running*.
- **Go 1.25+**, hanya diperlukan kalau app dijalankan di host atau untuk menjalankan test dan `make swag`. Di macOS: `brew install go`.
- **golangci-lint v2** (opsional, untuk `make lint`). Di macOS: `brew install golangci-lint`.

> Semua perintah di bawah dijalankan dari root folder project (`pokemon-service/`).

## Quick Start

### Opsi A: semua di Docker (paling cepat)

Menjalankan Postgres, migration, dan app sekaligus:

```bash
make up
```

Isi data contoh (6 pokemon):

```bash
make seed
```

Cek:

```bash
curl http://localhost:8080/health/ready
```

Buka Swagger UI: http://localhost:8080/swagger/index.html

### Opsi B: app di host, Postgres di Docker (untuk development)

```bash
cp .env.example .env
```

```bash
make db-up && make migrate-up && make seed
```

```bash
make run
```

### Menghentikan

```bash
make down
```

Data Postgres tersimpan di volume `pgdata`. Untuk menghapus data sekalian, jalankan `docker compose down -v`.

## Konfigurasi

Semua konfigurasi dibaca dari environment variable lewat [config/config.go](config/config.go). Contohnya ada di [.env.example](.env.example).

| Variable | Default | Keterangan |
|---|---|---|
| `APP_ENV` | `development` | `production` mengaktifkan Gin release mode dan mematikan Swagger |
| `HTTP_PORT` | `8080` | Port HTTP server |
| `LOG_LEVEL` | `info` | `debug` / `info` / `warn` / `error` |
| `SWAGGER_ENABLED` | `true` (kecuali production) | Paksa aktif/nonaktif Swagger UI |
| `DB_HOST` | `localhost` | Host Postgres (`postgres` di dalam Docker Compose) |
| `DB_PORT` | `5432` | |
| `DB_USER` | `postgres` | |
| `DB_PASSWORD` | `postgres` | |
| `DB_NAME` | `pokemon` | |
| `DB_SSLMODE` | `disable` | |
| `DB_MAX_OPEN_CONNS` | `25` | Connection pool |
| `DB_MAX_IDLE_CONNS` | `5` | Connection pool |
| `DB_CONN_MAX_LIFETIME` | `30m` | Format durasi Go (`30m`, `1h`) |

> `.env` sudah ada di `.gitignore`, jadi jangan commit secret.

## Perintah Makefile

| Perintah | Fungsi |
|---|---|
| `make up` | Build & jalankan Postgres + migration + app di Docker |
| `make down` | Hentikan semua container |
| `make logs` | Ikuti log app (`Ctrl+C` untuk keluar, container tetap jalan) |
| `make seed` | Isi data contoh dari [scripts/seed.sql](scripts/seed.sql) |
| `make db-up` | Jalankan Postgres saja |
| `make migrate-up` | Jalankan migration |
| `make migrate-down` | Rollback 1 migration terakhir |
| `make run` | Jalankan app di host (`go run ./cmd`) |
| `make build` | Build binary ke `bin/pokemon-service` |
| `make test` | Unit test (use case + handler) dengan race detector |
| `make test-integration` | Integration test repository ke Postgres (butuh `make db-up`) |
| `make lint` | golangci-lint |
| `make cover` | Unit test + laporan coverage |
| `make swag` | Generate ulang dokumentasi Swagger ke `docs/` |
| `make tidy` | `go mod tidy` |

## Dokumentasi API (Swagger)

- **Swagger UI:** http://localhost:8080/swagger/index.html (`/swagger` otomatis redirect ke sini)
- **Spec OpenAPI (JSON):** http://localhost:8080/swagger/doc.json

Dokumentasi di-generate dari anotasi `godoc` di [pokemon_handler.go](internal/delivery/http/handler/pokemon_handler.go), dengan info umum API di [cmd/main.go](cmd/main.go). Setiap kali anotasi atau DTO berubah:

```bash
make swag && docker compose up -d --build app
```

> Folder `docs/` **wajib di-commit**, karena Docker build tidak menjalankan `swag`. Jangan edit file di `docs/` secara manual.

## Endpoint

Base URL: `http://localhost:8080/api/v1`

| Method | Path | Keterangan | Sukses |
|---|---|---|---|
| GET | `/health/live` *(tanpa prefix)* | Liveness: proses hidup, tidak cek dependency | 200 |
| GET | `/health/ready` *(tanpa prefix)* | Readiness: ping database (timeout 2 detik) | 200 / 503 |
| GET | `/health` *(tanpa prefix)* | Alias `/health/ready` | 200 / 503 |
| GET | `/pokemons` | List + pagination + filter | 200 |
| GET | `/pokemons/{id}` | Detail | 200 |
| POST | `/pokemons` | Create | 201 |
| PATCH | `/pokemons/{id}` | Partial update | 200 |
| DELETE | `/pokemons/{id}` | Soft delete | 204 |

**Query param list:**

| Param | Default | Keterangan |
|---|---|---|
| `page` | `1` | Minimal 1 |
| `page_size` | `10` | 1–100 |
| `type` | - | Cocokkan dengan `primary_type` **atau** `secondary_type` |
| `search` | - | Cari nama (case-insensitive, partial match) |

**Body create/update:**

| Field | Tipe | Create | Keterangan |
|---|---|---|---|
| `pokedex_number` | int | wajib | ≥ 1, unik |
| `name` | string | wajib | 1–50 karakter, unik, disimpan lowercase |
| `primary_type` | string | wajib | Salah satu dari 18 tipe |
| `secondary_type` | string | opsional | Salah satu dari 18 tipe; pada PATCH kirim `""` untuk menghapus |
| `hp`, `attack`, `defense`, `speed` | int | wajib | 1–255 |

Pada PATCH semua field opsional, dan hanya field yang dikirim yang diubah.

Tipe yang valid: `normal`, `fire`, `water`, `grass`, `electric`, `ice`, `fighting`, `poison`, `ground`, `flying`, `psychic`, `bug`, `rock`, `ghost`, `dragon`, `dark`, `steel`, `fairy`.

## Contoh Request (curl / Postman)

**Import ke Postman:** pilih *Import → Link*, lalu masukkan `http://localhost:8080/swagger/doc.json`. Semua endpoint langsung jadi satu collection. Curl di bawah juga bisa di-import satu per satu lewat *Import → Raw text*.

> Ganti `{id}` dengan UUID pokemon. Ambil dari response list atau create.

List (halaman 2, 3 item per halaman):

```bash
curl 'http://localhost:8080/api/v1/pokemons?page=2&page_size=3'
```

List dengan filter tipe + search:

```bash
curl 'http://localhost:8080/api/v1/pokemons?type=fire&search=char'
```

Detail:

```bash
curl 'http://localhost:8080/api/v1/pokemons/{id}'
```

Create:

```bash
curl -X POST 'http://localhost:8080/api/v1/pokemons' -H 'Content-Type: application/json' -d '{"pokedex_number": 149, "name": "Dragonite", "primary_type": "dragon", "secondary_type": "flying", "hp": 91, "attack": 134, "defense": 95, "speed": 80}'
```

Partial update:

```bash
curl -X PATCH 'http://localhost:8080/api/v1/pokemons/{id}' -H 'Content-Type: application/json' -d '{"speed": 100, "secondary_type": "fairy"}'
```

Hapus secondary type:

```bash
curl -X PATCH 'http://localhost:8080/api/v1/pokemons/{id}' -H 'Content-Type: application/json' -d '{"secondary_type": ""}'
```

Delete:

```bash
curl -X DELETE 'http://localhost:8080/api/v1/pokemons/{id}'
```

Tip: di terminal, tambahkan `-s ... | python3 -m json.tool` supaya JSON lebih mudah dibaca.

Contoh response readiness saat database mati (`503`):

```json
{ "status": "unavailable", "checks": { "database": "down" } }
```

Gunakan `/health/live` untuk liveness probe (restart container) dan `/health/ready` untuk readiness probe (keluarkan dari load balancer), supaya DB yang sedang down tidak membuat container di-restart terus-menerus.

## Format Response & Error

Sukses (list):

```json
{
  "success": true,
  "data": [
    { "id": "cf11f1e0-...", "pokedex_number": 25, "name": "pikachu", "primary_type": "electric", "secondary_type": null, "hp": 35, "attack": 55, "defense": 40, "speed": 90, "created_at": "...", "updated_at": "..." }
  ],
  "meta": { "page": 1, "page_size": 10, "total_items": 7, "total_pages": 1 }
}
```

Sukses (detail/create/update): sama, tanpa `meta`, dan `data` berupa object. DELETE mengembalikan `204` tanpa body.

Error:

```json
{ "success": false, "error": { "code": "CONFLICT", "message": "pokedex number 25 sudah terdaftar" } }
```

| HTTP | `error.code` | Kapan |
|---|---|---|
| 400 | `BAD_REQUEST` | Body/query tidak valid, tipe tidak dikenal, stat di luar 1–255, ID bukan UUID |
| 404 | `NOT_FOUND` | Pokemon tidak ada (atau sudah dihapus) |
| 409 | `CONFLICT` | Nama atau `pokedex_number` sudah dipakai |
| 500 | `INTERNAL_ERROR` | Error tak terduga; detail hanya dicatat di log, tidak dikirim ke client |

## Aturan Bisnis

- `name` disimpan **lowercase** dan unik, sehingga `Pikachu` dan `pikachu` dianggap sama.
- `pokedex_number` unik.
- `secondary_type` tidak boleh sama dengan `primary_type`.
- Stat (`hp`, `attack`, `defense`, `speed`) di rentang 1–255. Rentang ini dicek di handler, use case, dan `CHECK` constraint DB.
- **Soft delete:** data tidak benar-benar dihapus (kolom `deleted_at` diisi). Unique index bersifat *partial* (`WHERE deleted_at IS NULL`), jadi nama dan nomor pokedex bisa dipakai lagi setelah dihapus.

## Arsitektur

```
cmd/main.go                         # entry point, DI manual, graceful shutdown
config/                             # konfigurasi dari env var
internal/
  domain/                           # tanpa dependency eksternal (gin/gorm)
    entity/                         # entity Pokemon + PokemonType
    repository/                     # interface PokemonRepository + filter
    usecase/                        # interface PokemonUseCase + request struct
  usecase/                          # business logic (validasi, cek duplikat) + unit test
  repository/postgres/              # implementasi GORM (model terpisah dari entity)
  delivery/http/
    handler/                        # Gin handler, DTO request/response, anotasi Swagger
    middleware/                     # request logger (slog)
    response/                       # format response & mapping AppError → HTTP status
    router/                         # registrasi route + Swagger UI
pkg/
  errors/                           # AppError (NOT_FOUND, CONFLICT, BAD_REQUEST)
  logger/                           # structured logger (log/slog)
  pagination/                       # helper offset pagination
docs/                               # OpenAPI hasil generate swag (jangan edit manual)
migrations/                         # file SQL golang-migrate
scripts/seed.sql                    # data contoh
```

Alur dependency:

```
handler ──▶ usecase ──▶ repository (interface di domain) ◀── postgres (implementasi GORM)
```

- Handler hanya bicara dengan use case, tidak pernah langsung ke repository.
- Dependency di-inject manual lewat constructor di [cmd/main.go](cmd/main.go).
- Semua fungsi IO menerima `context.Context`, dan error di-wrap dengan `fmt.Errorf("...: %w", err)`.

**Menambah resource baru** (misalnya `trainer`): buat entity dan interface di `internal/domain/`, implementasi use case di `internal/usecase/`, repository di `internal/repository/postgres/`, handler di `internal/delivery/http/handler/`, migration baru di `migrations/`, lalu daftarkan di [router.go](internal/delivery/http/router/router.go) dan [main.go](cmd/main.go).

## Testing

| Jenis | Lokasi | Butuh DB? | Perintah |
|---|---|---|---|
| Unit test use case | [pokemon_usecase_test.go](internal/usecase/pokemon_usecase_test.go) | Tidak (fake repository in-memory) | `make test` |
| Unit test HTTP handler | [pokemon_handler_test.go](internal/delivery/http/handler/pokemon_handler_test.go), [health_handler_test.go](internal/delivery/http/handler/health_handler_test.go) | Tidak (`httptest` + fake use case) | `make test` |
| Integration test repository | [pokemon_repository_integration_test.go](internal/repository/postgres/pokemon_repository_integration_test.go) | Ya (Postgres) | `make test-integration` |

**Yang dicakup:**
- *Use case:* validasi, conflict, not found, partial update, delete, normalisasi pagination, error wrapping.
- *Handler:* binding & validasi request, mapping `AppError` ke status HTTP, error 500 tidak membocorkan detail internal, `secondary_type: ""` menjadi hapus, health live/ready.
- *Repository:* query SQL sungguhan, unique index (conflict), filter type ke secondary, search case-insensitive, urutan & pagination, update ke `NULL`, soft delete, serta pemakaian ulang nama setelah dihapus.

Menjalankan integration test:

```bash
make db-up && make test-integration
```

Integration test diberi build tag `integration`, jadi tidak ikut `make test`. Tiap test membuat **schema Postgres sementara**, menjalankan semua file di `migrations/`, lalu menghapus schema itu, sehingga data di database lokal tidak tersentuh. DSN bisa diganti lewat env `TEST_DATABASE_DSN`. Kalau env ini kosong, test di-skip.

## CI (GitHub Actions)

Workflow ada di [`.github/workflows/pokemon-service.yml`](../.github/workflows/pokemon-service.yml) (di root repo) dan hanya berjalan kalau ada perubahan di `pokemon-service/**`.

| Job | Isi |
|---|---|
| `lint` | golangci-lint v2 (config di [.golangci.yml](.golangci.yml)) |
| `test` | Unit test + race detector + coverage |
| `integration` | Integration test repository dengan service container Postgres 16 |
| `swagger` | Jalankan `make swag`, gagal kalau `docs/` belum di-generate ulang dan di-commit |
| `docker` | Build image Docker (setelah `lint` dan `test` lolos) |

## Troubleshooting

| Gejala | Penyebab & solusi |
|---|---|
| `make: *** No rule to make target ...` | Perintah dijalankan di luar folder project. `cd` ke `pokemon-service/` dulu, atau pakai `make -C <path> <target>`. |
| `no configuration file provided: not found` | Sama: `docker compose` dijalankan di luar folder project. |
| `Cannot connect to the Docker daemon` | Docker Desktop belum jalan. Buka dan tunggu sampai statusnya *running*. |
| Perubahan Swagger tidak muncul | Jalankan `make swag`, lalu rebuild app: `docker compose up -d --build app`. |
| Port 5432 / 8080 sudah dipakai | Hentikan service lain di port tersebut, atau ubah mapping `ports` di [docker-compose.yml](docker-compose.yml). |
| `make test-integration` gagal `connection refused` | Postgres belum jalan. Jalankan `make db-up` dulu. |
| Job CI `swagger` gagal | Anotasi atau DTO berubah tapi `docs/` belum di-update. Jalankan `make swag`, lalu commit. |
| `go: command not found` | Install Go (`brew install go`), atau pakai Opsi A yang semuanya di Docker. |

## Langkah Lanjut

- Cache Redis di use case (interface `PokemonCache`, implementasi di `internal/repository/redis`).
- Middleware auth (JWT) dan rate limit.
- Deploy otomatis (push image ke registry) dari CI.
