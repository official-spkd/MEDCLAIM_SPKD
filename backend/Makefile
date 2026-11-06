# MedClaim — perintah singkat via make.
#   make up          : full stack Docker (db + backend + frontend), instance 1
#   make up I=2      : instance 2 (port otomatis: API 7082, DB 7434, FE 7042)
#   make down        : stop instance
#   make logs        : ikuti log
# Daftar lengkap: make help
#
# Resep ditulis agar jalan baik lewat Git Bash (sh) maupun PowerShell (cmd.exe),
# jadi jangan pakai sintaks khusus salah satu shell di sini.

# Blok port MedClaim = 7xxx, jauh dari sistem lain di mesin ini
# (simrs 8080/3000/8000/5432, medpath 8082/5434/5173, matpad 303x, dst).
I ?= 1
API_PORT ?= $(word $(I),7081 7082 7083 7084 7085 7086 7087 7088 7089)
DB_PORT  ?= $(word $(I),7433 7434 7435 7436 7437 7438 7439 7440 7441)
FE_PORT  ?= $(word $(I),7041 7042 7043 7044 7045 7046 7047 7048 7049)
PROJECT  := medclaim-i$(I)

# Diturunkan ke docker compose (port host) dan ke go run (DB per instance).
export MEDCLAIM_API_PORT = $(API_PORT)
export MEDCLAIM_DB_PORT  = $(DB_PORT)
export MEDCLAIM_FE_PORT  = $(FE_PORT)
export DATABASE_URL      = postgres://medclaim:medclaim@localhost:$(DB_PORT)/medclaim?sslmode=disable
export PORT              = $(API_PORT)
export CORS_ORIGINS      = http://localhost:$(FE_PORT),http://localhost:5173

COMPOSE := docker compose -p $(PROJECT)

.DEFAULT_GOAL := help
.PHONY: help up down clean restart rebuild logs ps db stop-db migrate fresh seed server test fe fe-demo

help:
	@echo MedClaim make targets - instance $(I): API $(API_PORT), DB $(DB_PORT), FE $(FE_PORT)
	@echo   make up         : jalankan full stack Docker, db + backend + frontend
	@echo   make down       : matikan instance ini
	@echo   make clean      : matikan DAN hapus data database
	@echo   make restart    : restart container
	@echo   make rebuild    : build ulang image lalu up
	@echo   make logs       : ikuti log semua service
	@echo   make ps         : status container
	@echo   make db         : nyalakan database saja, untuk mode dev go run
	@echo   make migrate    : go run ./cmd/migrate ke DB instance ini
	@echo   make fresh      : migrate -fresh, reset schema
	@echo   make seed       : go run ./cmd/seed
	@echo   make server     : go run ./cmd/server di port $(API_PORT)
	@echo   make fe         : frontend Vite dev di :$(FE_PORT), menunjuk API :$(API_PORT)
	@echo   make fe-demo    : frontend Vite dev mode demo WASM, tanpa backend
	@echo   make test       : go test ./...
	@echo Instance lain: tambah I=2, I=3, dst pada perintah apa pun.

up:
	$(COMPOSE) up -d --build
	@echo Frontend: http://localhost:$(FE_PORT)/app/
	@echo API:      http://localhost:$(API_PORT)

down:
	$(COMPOSE) down

# Hati-hati: menghapus volume data postgres instance ini.
clean:
	$(COMPOSE) down -v

restart:
	$(COMPOSE) restart

rebuild:
	$(COMPOSE) build --no-cache
	$(COMPOSE) up -d

logs:
	$(COMPOSE) logs -f --tail=100

ps:
	$(COMPOSE) ps

db:
	$(COMPOSE) up -d db

stop-db:
	$(COMPOSE) stop db

migrate:
	go run ./cmd/migrate

fresh:
	go run ./cmd/migrate -fresh

seed:
	go run ./cmd/seed

server:
	go run ./cmd/server

test:
	go test ./...

# --- Frontend dev (Vite hot-reload) ---
# Port sama dengan FE Docker ($(FE_PORT)); matikan container frontend dulu bila bentrok
# (make down, atau: docker compose -p $(PROJECT) stop frontend).
# Env OS menang atas file .env* Vite, jadi VITE_API_BASE/VITE_MOCK diatur per target di sini.

fe: export VITE_API_BASE = http://localhost:$(API_PORT)
fe: export VITE_MOCK =
fe:
ifeq ($(wildcard ../frontend/node_modules),)
	cd ../frontend && npm install
endif
	@echo Frontend dev http://localhost:$(FE_PORT)/app/ memakai API http://localhost:$(API_PORT)
	cd ../frontend && npx vite --port $(FE_PORT) --host --strictPort

fe-demo: export VITE_API_BASE =
fe-demo: export VITE_MOCK = 1
fe-demo:
ifeq ($(wildcard ../frontend/node_modules),)
	cd ../frontend && npm install
endif
	@echo Frontend dev MODE DEMO - WASM data dummy: http://localhost:$(FE_PORT)/app/
	cd ../frontend && npx vite --port $(FE_PORT) --host --strictPort
