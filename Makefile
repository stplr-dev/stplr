NAME := stplr
GIT_VERSION ?= $(shell git describe --tags )
IGNORE_ROOT_CHECK ?= 0
DESTDIR ?=

CONFIG_DEFAULT_ROOT_CMD ?= 

PREFIX ?= /usr/local
datarootdir = $(PREFIX)/share
datadir = $(datarootdir)
exec_prefix = $(PREFIX)
bindir = $(exec_prefix)/bin
sysconfdir = /etc
mandir = $(datarootdir)/man

BIN := ./$(NAME)
COMPLETIONS_DIR := ./scripts/completion
BASH_COMPLETION := $(COMPLETIONS_DIR)/bash
ZSH_COMPLETION := $(COMPLETIONS_DIR)/zsh
MAN_DIR := ./man

GENERATE ?= 1
POST_INSTALL ?= 1

CACHE_DIR ?= /var/cache/stplr
SYSUSERS_DIR ?= /usr/lib/sysusers.d
TMPFILES_DIR ?= /usr/lib/tmpfiles.d
SYSUSERS_CONF := packaging/stplr.sysusers
TMPFILES_CONF := packaging/stplr.tmpfiles
DEFAULT_CONF := packaging/stplr.toml
DEFAULT_FIREJAILED_CONF := packaging/firejailed/global

ADD_LICENSE_BIN := go run github.com/google/addlicense@4caba19b7ed7818bb86bc4cd20411a246aa4a524
GOLANGCI_LINT_BIN := go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.3.1
XGOTEXT_BIN := go run github.com/Tom5521/xgotext@v1.2.0

LDFLAGS := -X 'go.stplr.dev/stplr/internal/config.Version=$(GIT_VERSION)'
ifneq ($(CONFIG_DEFAULT_ROOT_CMD),)
	LDFLAGS += -X 'go.stplr.dev/stplr/internal/constants.ConfigDefaultRootCmd=$(CONFIG_DEFAULT_ROOT_CMD)'
endif

.PHONY: build install clean clear uninstall install-config install-cachedir install-sysusers install-tmpfiles install-post

build: $(BIN) $(MAN_DIR)

$(BIN):
ifeq ($(GENERATE),1)
	go generate ./...
else
	@echo "Skipping go generate (GENERATE=0)"
endif
	go build -ldflags="$(LDFLAGS)" -o $@ ./cmd/stplr

$(MAN_DIR):
	@echo "Generating man pages..."
	@mkdir -p $(MAN_DIR)/en/man1
	@mkdir -p $(MAN_DIR)/ru/man1
	@LANG=en_US.UTF-8 go run ./tools/gen-docs/main.go > $(MAN_DIR)/en/man1/$(NAME).1
	@LANG=ru_RU.UTF-8 go run ./tools/gen-docs/main.go > $(MAN_DIR)/ru/man1/$(NAME).1
	@echo "Man pages generated in $(MAN_DIR)/"

install-man: $(MAN_DIR)
	@echo "Installing man pages..."
	install -Dm644 $(MAN_DIR)/en/man1/$(NAME).1 $(DESTDIR)$(mandir)/man1/$(NAME).1
	install -Dm644 $(MAN_DIR)/ru/man1/$(NAME).1 $(DESTDIR)$(mandir)/ru/man1/$(NAME).1

install: build install-config install-sysusers install-tmpfiles install-cachedir install-man
	install -Dm755 $(BIN) $(DESTDIR)$(bindir)/$(NAME)
	
	@mkdir -p $(DESTDIR)$(datadir)/bash-completion/completions
	@$(BIN) completion bash > $(DESTDIR)$(datadir)/bash-completion/completions/$(NAME)
	
	@mkdir -p $(DESTDIR)$(datadir)/zsh/site-functions
	@$(BIN) completion zsh > $(DESTDIR)$(datadir)/zsh/site-functions/_$(NAME)

	install -Dm755 packaging/stplr.fish $(DESTDIR)$(datadir)/fish/vendor_completions.d/$(NAME).fish

ifeq ($(POST_INSTALL),1)
	$(MAKE) install-post
endif	
	@echo "Installation done!"

install-config:
	install -d -m 755 $(DESTDIR)$(sysconfdir)/stplr
	[ -f $(DESTDIR)$(sysconfdir)/stplr/stplr.toml ] || install -m 644 $(DEFAULT_CONF) $(DESTDIR)$(sysconfdir)/stplr/stplr.toml
	install -d -m 755 $(DESTDIR)$(sysconfdir)/stplr/firejailed
	[ -f $(DESTDIR)$(sysconfdir)/stplr/firejailed/global ] || install -m 644 $(DEFAULT_FIREJAILED_CONF) $(DESTDIR)$(sysconfdir)/stplr/firejailed/global

install-sysusers:
	install -Dpm644 $(SYSUSERS_CONF) $(DESTDIR)$(SYSUSERS_DIR)/stplr.conf

install-tmpfiles:
	install -Dpm644 $(TMPFILES_CONF) $(DESTDIR)$(TMPFILES_DIR)/stplr.conf

install-cachedir:
	install -d -m 755 $(DESTDIR)$(CACHE_DIR)

install-post:
	@echo "Running post-installation system setup..."
	@if ! id stapler-builder >/dev/null 2>&1; then \
		useradd -r -s /usr/sbin/nologin stapler-builder; \
	else \
		echo "User 'stapler-builder' already exists. Skipping."; \
	fi
	install -d -o stapler-builder -g stapler-builder -m 755 $(DESTDIR)$(CACHE_DIR);

uninstall:
	rm -rf \
		$(DESTDIR)$(bindir)/$(NAME) \
		$(DESTDIR)$(datadir)/bash-completion/completions/$(NAME) \
		$(DESTDIR)$(datadir)/zsh/site-functions/_$(NAME) \
		$(DESTDIR)$(datadir)/fish/vendor_completions.d/$(NAME).fish \
		$(DESTDIR)$(mandir)/man1/$(NAME).1 \
		$(DESTDIR)$(mandir)/ru/man1/$(NAME).1

clean clear:
	rm -f $(BIN)
	rm -rf $(MAN_DIR)

# Development Targets

fmt:
	$(GOLANGCI_LINT_BIN) run --fix	

i18n:
	$(XGOTEXT_BIN) --output ./internal/i18n/default.pot
	msguniq --use-first -o ./internal/i18n/default.pot ./internal/i18n/default.pot
	msgmerge --backup=off -U ./internal/i18n/po/ru/default.po ./internal/i18n/default.pot

test: test-unit-coverage test-e2e
	@echo "All tests completed successfully!"

test-unit:
	go test ./... -v

test-unit-coverage:
	go test ./... -v -coverpkg=./... -coverprofile=coverage.out
	bash scripts/coverage-badge.sh

prepare-test-e2e: clean build
	rm -f e2e-tests/$(NAME)
	cp $(NAME) e2e-tests

test-e2e: prepare-test-e2e
	go test -v -timeout 10m -parallel 4 -tags=e2e ./e2e-tests/...

update-license:
	$(ADD_LICENSE_BIN) -v -ignore 'packaging/**' -ignore 'vendor/**' -f license-header.tmpl .

update-deps-cve:
	bash scripts/update-deps-cve.sh

mocks: \
	internal/service/repos/internal/gitmanager/gitmanager.go \
	internal/build/utils.go \
	internal/usecase/support/archive.go \
	internal/usecase/repo/list/list.go \
	internal/usecase/config/get/get.go \
	internal/service/updater/updater.go
	@echo "Generating mocks..."
	@for file in $^; do \
		mockgen -source=$$file -destination=$$(dirname $$file)/mock_$$(basename $$file .go)_test.go -package=$$(basename $$(dirname $$file)); \
	done

create-release-artifacts: clean build
	bash scripts/create-release-artifacts.sh
