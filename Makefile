NAME := stplr
GIT_VERSION ?= $(shell git describe --tags )
IGNORE_ROOT_CHECK ?= 0
DESTDIR ?=
PREFIX ?= /usr/local
BIN := ./$(NAME)
INSTALLED_BIN := $(DESTDIR)/$(PREFIX)/bin/$(NAME)
COMPLETIONS_DIR := ./scripts/completion
BASH_COMPLETION := $(COMPLETIONS_DIR)/bash
ZSH_COMPLETION := $(COMPLETIONS_DIR)/zsh
INSTALLED_BASH_COMPLETION := $(DESTDIR)$(PREFIX)/share/bash-completion/completions/$(NAME)
INSTALLED_ZSH_COMPLETION := $(DESTDIR)$(PREFIX)/share/zsh/site-functions/_$(NAME)

GENERATE ?= 1

CREATE_SYSTEM_RESOURCES ?= 1
ROOT_DIRS := /var/cache/stplr /etc/stplr

ADD_LICENSE_BIN := go run github.com/google/addlicense@4caba19b7ed7818bb86bc4cd20411a246aa4a524
GOLANGCI_LINT_BIN := go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.63.4
XGOTEXT_BIN := go run github.com/Tom5521/xgotext@v1.2.0

.PHONY: build install clean clear uninstall check-no-root

build: check-no-root $(BIN)

export CGO_ENABLED := 0
$(BIN):
ifeq ($(GENERATE),1)
	go generate ./...
else
	@echo "Skipping go generate (GENERATE=0)"
endif
	go build -ldflags="-X 'go.stplr.dev/stplr/internal/config.Version=$(GIT_VERSION)'" -o $@

check-no-root:
	@if [ "$$IGNORE_ROOT_CHECK" != "1" ] && [ "`whoami`" = "root" ]; then \
		echo "This target shouldn't run as root" 1>&2; \
		echo "Set IGNORE_ROOT_CHECK=1 to override" 1>&2; \
		exit 1; \
	fi

install: \
	$(INSTALLED_BIN) \
	$(INSTALLED_BASH_COMPLETION) \
	$(INSTALLED_ZSH_COMPLETION)
	@echo "Installation done!"

$(INSTALLED_BIN): $(BIN)
	install -Dm755 $< $@
ifeq ($(CREATE_SYSTEM_RESOURCES),1)
	setcap cap_setuid,cap_setgid+ep $(INSTALLED_BIN)
	@if id stapler-builder >/dev/null 2>&1; then \
		echo "User 'stapler-builder' already exists. Skipping."; \
	else \
		useradd -r -s /usr/sbin/nologin stapler-builder; \
	fi
	@for dir in $(ROOT_DIRS); do \
		install -d -o stapler-builder -g stapler-builder -m 755 $$dir; \
	done
else
	@echo "Skipping user and root dir creation (CREATE_SYSTEM_RESOURCES=0)"
endif

$(INSTALLED_BASH_COMPLETION): $(BASH_COMPLETION)
	install -Dm755 $< $@

$(INSTALLED_ZSH_COMPLETION): $(ZSH_COMPLETION)
	install -Dm755 $< $@

uninstall:
	rm -f \
		$(INSTALLED_BIN) \
		$(INSTALLED_BASH_COMPLETION) \
		$(INSTALLED_ZSH_COMPLETION)

clean clear:
	rm -f $(BIN)

OLD_FILES=$(shell cat old-files)
IGNORE_OLD_FILES := $(foreach file,$(shell cat old-files),-ignore $(file))
update-license:
	$(ADD_LICENSE_BIN) -v -f license-header.tmpl $(IGNORE_OLD_FILES) .

fmt:
	$(GOLANGCI_LINT_BIN) run --fix

i18n:
	$(XGOTEXT_BIN)  --output ./internal/translations/default.pot
	msguniq --use-first -o ./internal/translations/default.pot ./internal/translations/default.pot
	msgmerge --backup=off -U ./internal/translations/po/ru/default.po ./internal/translations/default.pot
	bash scripts/i18n-badge.sh

test-coverage:
	go test ./... -v -coverpkg=./... -coverprofile=coverage.out
	bash scripts/coverage-badge.sh

update-deps-cve:
	bash scripts/update-deps-cve.sh

prepare-for-e2e-test: clean build
	rm -f ./e2e-tests/alr
	cp $(NAME) e2e-tests

e2e-test: prepare-for-e2e-test
	go test -v -timeout 10m -tags=e2e ./...