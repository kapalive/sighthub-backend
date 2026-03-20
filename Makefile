# ──────────────────────────────────────────────────────────────────────────────
# sighthub-backend  Makefile
# ──────────────────────────────────────────────────────────────────────────────
SHELL        := /bin/bash
APP_NAME     := sighthub-backend
APP_DIR      := $(shell pwd)
BINARY       := $(APP_DIR)/$(APP_NAME)
SERVICE_USER := $(APP_NAME)
SERVICE_GROUP:= $(APP_NAME)
GO_VERSION   := 1.24.0
SECRETS_DIR  := $(APP_DIR)/secrets
CONFIG_DIR   := $(APP_DIR)/config

# Определяем окружение (по умолчанию production)
APP_ENV      ?= production
ifeq ($(APP_ENV),development)
  CONFIG_FILE := $(CONFIG_DIR)/config.development.json
  SERVICE_FILE := $(APP_NAME)-dev
  PORT        := 8002
else
  CONFIG_FILE := $(CONFIG_DIR)/config.json
  SERVICE_FILE := $(APP_NAME)-prod
  PORT        := 8001
endif

.PHONY: help check check-os check-go check-redis check-jq \
        install-go install-redis install-jq install-deps \
        deps build clean \
        setup-user setup-secrets setup-config setup-service \
        install \
        start stop restart status logs \
        uninstall

# ─── Help ─────────────────────────────────────────────────────────────────────

help:
	@echo ""
	@echo "  $(APP_NAME) Makefile"
	@echo "  ─────────────────────────────────────────"
	@echo ""
	@echo "  Проверка и установка зависимостей:"
	@echo "    make check            — проверить Go, Redis, jq, OS"
	@echo "    make install-go       — установить Go $(GO_VERSION)"
	@echo "    make install-redis    — установить Redis"
	@echo "    make install-deps     — установить всё недостающее (Go, Redis, jq)"
	@echo ""
	@echo "  Сборка:"
	@echo "    make deps             — go mod download"
	@echo "    make build            — собрать бинарник"
	@echo "    make clean            — удалить бинарник"
	@echo ""
	@echo "  Установка сервиса:"
	@echo "    make setup-user       — создать пользователя/группу $(SERVICE_GROUP)"
	@echo "    make setup-secrets    — сгенерировать RSA ключи и секреты (если нет)"
	@echo "    make setup-config     — проставить секреты в конфиг (если пустые)"
	@echo "    make setup-service    — установить systemd unit"
	@echo "    make install          — полная установка (всё выше)"
	@echo ""
	@echo "  Управление сервисом:"
	@echo "    make start            — запустить"
	@echo "    make stop             — остановить"
	@echo "    make restart          — перезапустить"
	@echo "    make status           — статус"
	@echo "    make logs             — journalctl -f"
	@echo ""
	@echo "  Переменные:"
	@echo "    APP_ENV=production    — prod (порт 8001, по умолчанию)"
	@echo "    APP_ENV=development   — dev  (порт 8002)"
	@echo ""

# ─── OS Detection ─────────────────────────────────────────────────────────────

OS_ID       := $(shell . /etc/os-release 2>/dev/null && echo $$ID)
OS_ID_LIKE  := $(shell . /etc/os-release 2>/dev/null && echo $$ID_LIKE)
OS_VERSION  := $(shell . /etc/os-release 2>/dev/null && echo $$VERSION_ID)
ARCH        := $(shell uname -m)

# Go binary — ищем и в PATH, и в /usr/local/go/bin
GO_BIN       := $(shell command -v go 2>/dev/null || echo /usr/local/go/bin/go)

# Нормализуем архитектуру для Go
ifeq ($(ARCH),x86_64)
  GO_ARCH := amd64
else ifeq ($(ARCH),aarch64)
  GO_ARCH := arm64
else
  GO_ARCH := $(ARCH)
endif

# Определяем пакетный менеджер
ifeq ($(OS_ID),ubuntu)
  PKG_MGR := apt-get
  PKG_UPDATE := sudo apt-get update -qq
  PKG_INSTALL := sudo apt-get install -y -qq
else ifeq ($(OS_ID),debian)
  PKG_MGR := apt-get
  PKG_UPDATE := sudo apt-get update -qq
  PKG_INSTALL := sudo apt-get install -y -qq
else ifneq (,$(findstring rhel,$(OS_ID_LIKE)))
  PKG_MGR := dnf
  PKG_UPDATE := true
  PKG_INSTALL := sudo dnf install -y -q
else ifneq (,$(findstring fedora,$(OS_ID)))
  PKG_MGR := dnf
  PKG_UPDATE := true
  PKG_INSTALL := sudo dnf install -y -q
else
  PKG_MGR := unknown
  PKG_UPDATE := true
  PKG_INSTALL := echo "ОШИБКА: неизвестный пакетный менеджер, установите вручную:"
endif

# ─── Checks ───────────────────────────────────────────────────────────────────

check: check-os check-go check-redis check-jq
	@echo ""
	@echo "✓ Все проверки пройдены"

check-os:
	@echo "── OS ──"
	@echo "  Дистрибутив : $(OS_ID) $(OS_VERSION)"
	@echo "  Архитектура : $(ARCH) (Go: $(GO_ARCH))"
	@echo "  Пакетный мгр: $(PKG_MGR)"

check-go:
	@echo "── Go ──"
	@if command -v go >/dev/null 2>&1; then \
		INSTALLED=$$(go version | grep -oP 'go\K[0-9]+\.[0-9]+'); \
		echo "  Установлен  : go$$INSTALLED (PATH)"; \
	elif [ -x /usr/local/go/bin/go ]; then \
		INSTALLED=$$(/usr/local/go/bin/go version | grep -oP 'go\K[0-9]+\.[0-9]+'); \
		echo "  Установлен  : go$$INSTALLED (/usr/local/go/bin)"; \
	else \
		echo "  ✗ Go не найден. Запустите: make install-go"; \
		exit 1; \
	fi

check-redis:
	@echo "── Redis ──"
	@if command -v redis-cli >/dev/null 2>&1; then \
		echo "  Установлен  : $$(redis-cli --version | head -1)"; \
		if redis-cli ping >/dev/null 2>&1; then \
			echo "  Статус      : работает"; \
		else \
			echo "  ⚠ Redis установлен, но не отвечает на PING"; \
		fi; \
	else \
		echo "  ✗ Redis не найден. Запустите: make install-redis"; \
		exit 1; \
	fi

check-jq:
	@echo "── jq ──"
	@if command -v jq >/dev/null 2>&1; then \
		echo "  Установлен  : $$(jq --version)"; \
	else \
		echo "  ✗ jq не найден. Запустите: make install-jq"; \
		exit 1; \
	fi

# ─── Install Prerequisites ───────────────────────────────────────────────────

install-go:
	@echo "── Установка Go $(GO_VERSION) ──"
	@if command -v go >/dev/null 2>&1; then \
		echo "  Go уже установлен: $$(go version)"; \
	elif [ -x /usr/local/go/bin/go ]; then \
		echo "  Go уже установлен: $$(/usr/local/go/bin/go version)"; \
	else \
		GO_TAR="go$(GO_VERSION).linux-$(GO_ARCH).tar.gz"; \
		echo "  Скачиваем $$GO_TAR ..."; \
		curl -fsSL "https://go.dev/dl/$$GO_TAR" -o "/tmp/$$GO_TAR"; \
		if [ ! -f "/tmp/$$GO_TAR" ]; then \
			echo "  ✗ Не удалось скачать $$GO_TAR"; \
			exit 1; \
		fi; \
		echo "  Распаковываем в /usr/local ..."; \
		sudo rm -rf /usr/local/go; \
		sudo tar -C /usr/local -xzf "/tmp/$$GO_TAR"; \
		rm -f "/tmp/$$GO_TAR"; \
		if ! grep -q '/usr/local/go/bin' /etc/profile.d/go.sh 2>/dev/null; then \
			echo 'export PATH=$$PATH:/usr/local/go/bin' | sudo tee /etc/profile.d/go.sh >/dev/null; \
		fi; \
		echo "  ✓ Установлен: $$(/usr/local/go/bin/go version)"; \
	fi

install-redis:
	@echo "── Установка Redis ──"
	@if command -v redis-cli >/dev/null 2>&1; then \
		echo "  Redis уже установлен"; \
	else \
		$(PKG_UPDATE); \
		$(PKG_INSTALL) redis-server redis-tools 2>/dev/null || $(PKG_INSTALL) redis; \
		sudo systemctl enable --now redis-server 2>/dev/null || sudo systemctl enable --now redis; \
		echo "  ✓ Redis установлен и запущен"; \
	fi

install-jq:
	@echo "── Установка jq ──"
	@if command -v jq >/dev/null 2>&1; then \
		echo "  jq уже установлен"; \
	else \
		$(PKG_UPDATE); \
		$(PKG_INSTALL) jq; \
		echo "  ✓ jq установлен"; \
	fi

install-pypdf:
	@echo "── Python3 + pypdf ──"
	@if python3 -c "import pypdf" 2>/dev/null; then \
		echo "  pypdf уже установлен"; \
	else \
		if ! command -v python3 >/dev/null 2>&1; then \
			$(PKG_UPDATE); \
			$(PKG_INSTALL) python3 python3-pip; \
		fi; \
		sudo pip3 install --break-system-packages pypdf 2>/dev/null \
			|| pip3 install --break-system-packages pypdf 2>/dev/null \
			|| pip3 install pypdf; \
		echo "  ✓ pypdf установлен"; \
	fi

install-deps: install-go install-redis install-jq install-pypdf
	@echo ""
	@echo "✓ Все зависимости установлены"

# ─── Build ────────────────────────────────────────────────────────────────────

GO_CMD := $(shell command -v go 2>/dev/null || echo /usr/local/go/bin/go)

deps:
	@echo "── Go mod download ──"
	$(GO_CMD) mod download
	$(GO_CMD) mod tidy

build: deps
	@echo "── Сборка $(APP_NAME) ──"
	@rm -f $(BINARY)
	CGO_ENABLED=0 $(GO_CMD) build -o $(BINARY) ./cmd/main.go
	@echo "  ✓ Бинарник: $(BINARY)"

clean:
	@echo "── Очистка ──"
	rm -f $(BINARY) main
	@echo "  ✓ Готово"

# ─── Setup User/Group ────────────────────────────────────────────────────────

setup-user:
	@echo "── Пользователь и группа $(SERVICE_GROUP) ──"
	@if getent group $(SERVICE_GROUP) >/dev/null 2>&1; then \
		echo "  Группа $(SERVICE_GROUP) уже существует"; \
	else \
		sudo groupadd --system $(SERVICE_GROUP); \
		echo "  ✓ Группа $(SERVICE_GROUP) создана"; \
	fi
	@if id $(SERVICE_USER) >/dev/null 2>&1; then \
		echo "  Пользователь $(SERVICE_USER) уже существует"; \
	else \
		sudo useradd --system --gid $(SERVICE_GROUP) --no-create-home \
			--shell /usr/sbin/nologin $(SERVICE_USER); \
		echo "  ✓ Пользователь $(SERVICE_USER) создан"; \
	fi
	@# Добавить текущего пользователя в группу сервиса
	@if ! id -nG $$USER 2>/dev/null | grep -qw $(SERVICE_GROUP); then \
		sudo usermod -aG $(SERVICE_GROUP) $$USER; \
		echo "  ✓ $$USER добавлен в группу $(SERVICE_GROUP)"; \
	fi

# ─── Setup Secrets ────────────────────────────────────────────────────────────

setup-secrets:
	@echo "── Генерация ключей и секретов ──"
	@mkdir -p $(SECRETS_DIR)
	@# RSA private key
	@if [ -f $(SECRETS_DIR)/client_sighthub_private.pem ]; then \
		echo "  Приватный ключ уже существует"; \
	else \
		openssl genrsa -out $(SECRETS_DIR)/client_sighthub_private.pem 2048 2>/dev/null; \
		echo "  ✓ Приватный ключ сгенерирован"; \
	fi
	@# RSA public key
	@if [ -f $(SECRETS_DIR)/client_sighthub_public.pem ]; then \
		echo "  Публичный ключ уже существует"; \
	else \
		openssl rsa -in $(SECRETS_DIR)/client_sighthub_private.pem \
			-pubout -out $(SECRETS_DIR)/client_sighthub_public.pem 2>/dev/null; \
		echo "  ✓ Публичный ключ сгенерирован"; \
	fi
	@# PKCS8 key
	@if [ -f $(SECRETS_DIR)/client_sighthub_private.pk8.pem ]; then \
		echo "  PKCS8 ключ уже существует"; \
	else \
		openssl pkcs8 -topk8 -nocrypt \
			-in $(SECRETS_DIR)/client_sighthub_private.pem \
			-out $(SECRETS_DIR)/client_sighthub_private.pk8.pem 2>/dev/null; \
		echo "  ✓ PKCS8 ключ сгенерирован"; \
	fi
	@# Права: владелец сервис, группа сервис, 640 на приватные, 644 на публичный
	@sudo chown $(SERVICE_USER):$(SERVICE_GROUP) $(SECRETS_DIR)/*.pem 2>/dev/null || true
	@chmod 640 $(SECRETS_DIR)/client_sighthub_private.pem \
		$(SECRETS_DIR)/client_sighthub_private.pk8.pem 2>/dev/null || true
	@chmod 644 $(SECRETS_DIR)/client_sighthub_public.pem 2>/dev/null || true
	@echo "  ✓ Права установлены (640 private, 644 public)"

# ─── Setup Config (fill missing secrets) ──────────────────────────────────────

setup-config:
	@echo "── Проверка секретов в конфиге ──"
	@if [ ! -f $(CONFIG_FILE) ]; then \
		echo "  ✗ Конфиг $(CONFIG_FILE) не найден"; \
		exit 1; \
	fi
	@CHANGED=0; \
	TMP=$$(mktemp); \
	cp $(CONFIG_FILE) $$TMP; \
	for KEY in secret_key jwt_secret_key refresh_secret_key; do \
		VAL=$$(jq -r ".$$KEY // empty" $$TMP); \
		if [ -z "$$VAL" ]; then \
			NEW=$$(openssl rand -hex 32); \
			jq --arg k "$$KEY" --arg v "$$NEW" '.[$k] = $v' $$TMP > $$TMP.new && mv $$TMP.new $$TMP; \
			echo "  ✓ $$KEY сгенерирован"; \
			CHANGED=1; \
		else \
			echo "  $$KEY уже заполнен"; \
		fi; \
	done; \
	PRIV=$$(jq -r '.private_key_path // empty' $$TMP); \
	if [ -z "$$PRIV" ]; then \
		jq '.private_key_path = "./secrets/client_sighthub_private.pem"' $$TMP > $$TMP.new && mv $$TMP.new $$TMP; \
		echo "  ✓ private_key_path установлен"; \
		CHANGED=1; \
	else \
		echo "  private_key_path уже заполнен"; \
	fi; \
	if [ "$$CHANGED" = "1" ]; then \
		cp $$TMP $(CONFIG_FILE); \
		echo "  ✓ Конфиг обновлён"; \
	fi; \
	rm -f $$TMP $$TMP.new
	@# Права на конфиг
	@sudo chown $(SERVICE_USER):$(SERVICE_GROUP) $(CONFIG_FILE) 2>/dev/null || true
	@chmod 640 $(CONFIG_FILE) 2>/dev/null || true

# ─── Setup Systemd Service ───────────────────────────────────────────────────

setup-service:
	@echo "── Установка systemd сервиса $(SERVICE_FILE) ──"
	@if [ ! -f system/$(SERVICE_FILE).service ]; then \
		echo "  ✗ Файл system/$(SERVICE_FILE).service не найден"; \
		exit 1; \
	fi
	@sudo cp system/$(SERVICE_FILE).service /etc/systemd/system/$(SERVICE_FILE).service
	@sudo systemctl daemon-reload
	@sudo systemctl enable $(SERVICE_FILE)
	@echo "  ✓ Сервис $(SERVICE_FILE) установлен и включён"

# ─── Full Install ─────────────────────────────────────────────────────────────

install: install-deps check setup-user build setup-secrets setup-config setup-service
	@echo ""
	@# Права: только бинарник, конфиги и секреты принадлежат сервису
	@sudo chown $(SERVICE_USER):$(SERVICE_GROUP) $(BINARY)
	@sudo chmod 750 $(BINARY)
	@sudo chown $(SERVICE_USER):$(SERVICE_GROUP) $(CONFIG_DIR)/config.json $(CONFIG_DIR)/config.development.json 2>/dev/null || true
	@sudo chmod 640 $(CONFIG_DIR)/config.json $(CONFIG_DIR)/config.development.json 2>/dev/null || true
	@sudo chown -R $(SERVICE_USER):$(SERVICE_GROUP) $(SECRETS_DIR)
	@# Рабочая директория должна быть читаема сервисом
	@sudo chmod o+rx $(APP_DIR)
	@echo "══════════════════════════════════════════"
	@echo "  ✓ $(APP_NAME) установлен"
	@echo "  Запуск: make start APP_ENV=$(APP_ENV)"
	@echo "══════════════════════════════════════════"

# ─── Update (build + systemd restart) ─────────────────────────────────────────

update: build
	@sudo chown $(SERVICE_USER):$(SERVICE_GROUP) $(BINARY)
	@sudo chmod 750 $(BINARY)
	@sudo systemctl daemon-reload
	@sudo systemctl enable $(APP_NAME)-prod $(APP_NAME)-dev 2>/dev/null || true
	@sudo systemctl restart $(APP_NAME)-prod
	@sudo systemctl restart $(APP_NAME)-dev
	@sleep 1
	@sudo systemctl is-active --quiet $(APP_NAME)-prod && echo "  ✓ prod (8001) запущен" || echo "  ✗ prod НЕ ПОДНЯЛСЯ"
	@sudo systemctl is-active --quiet $(APP_NAME)-dev  && echo "  ✓ dev  (8002) запущен" || echo "  ✗ dev НЕ ПОДНЯЛСЯ"
	@echo "✓ $(APP_NAME) обновлён и перезапущен"

# ─── Service Control ──────────────────────────────────────────────────────────

start:
	@sudo systemctl start $(SERVICE_FILE)
	@echo "✓ $(SERVICE_FILE) запущен"

stop:
	@sudo systemctl stop $(SERVICE_FILE)
	@echo "✓ $(SERVICE_FILE) остановлен"

restart: build
	@sudo systemctl restart $(SERVICE_FILE)
	@echo "✓ $(SERVICE_FILE) перезапущен"

status:
	@sudo systemctl status $(SERVICE_FILE) --no-pager

logs:
	@sudo journalctl -u $(SERVICE_FILE) -f --no-pager

# ─── Uninstall ────────────────────────────────────────────────────────────────

uninstall:
	@echo "── Удаление сервиса $(SERVICE_FILE) ──"
	@sudo systemctl stop $(SERVICE_FILE) 2>/dev/null || true
	@sudo systemctl disable $(SERVICE_FILE) 2>/dev/null || true
	@sudo rm -f /etc/systemd/system/$(SERVICE_FILE).service
	@sudo systemctl daemon-reload
	@echo "  ✓ Сервис удалён"
