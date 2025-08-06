NAME := stplr
GIT_VERSION ?= $(shell git describe --tags )
IGNORE_ROOT_CHECK ?= 0
DESTDIR ?=

prefix = /usr/local
datarootdir = $(prefix)/share
datadir = $(datarootdir)
exec_prefix = $(prefix)
bindir = $(exec_prefix)/bin
sysconfdir = /etc

BIN := ./$(NAME)
COMPLETIONS_DIR := ./scripts/completion
BASH_COMPLETION := $(COMPLETIONS_DIR)/bash
ZSH_COMPLETION := $(COMPLETIONS_DIR)/zsh

GENERATE ?= 1
POST_INSTALL ?= 1

CACHE_DIR ?= /var/cache/stplr
SYSUSERS_DIR ?= /usr/lib/sysusers.d
TMPFILES_DIR ?= /usr/lib/tmpfiles.d
SYSUSERS_CONF := packaging/stplr.sysusers
TMPFILES_CONF := packaging/stplr.tmpfiles
DEFAULT_CONF := packaging/stplr.toml

ADD_LICENSE_BIN := go run github.com/google/addlicense@4caba19b7ed7818bb86bc4cd20411a246aa4a524
GOLANGCI_LINT_BIN := go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.63.4
XGOTEXT_BIN := go run github.com/Tom5521/xgotext@v1.2.0

.PHONY: build install clean clear uninstall check-no-root install-config install-cachedir install-sysusers install-tmpfiles install-post

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

install: build install-config install-sysusers install-tmpfiles install-cachedir
	install -Dm755 $(BIN) $(DESTDIR)$(bindir)/$(NAME)
	install -Dm755 $(BASH_COMPLETION) $(DESTDIR)$(datadir)/bash-completion/completions/$(NAME)
	install -Dm755 $(ZSH_COMPLETION) $(DESTDIR)$(datadir)/zsh/site-functions/_$(NAME)
ifeq ($(POST_INSTALL),1)
	$(MAKE) install-post
endif	
	@echo "Installation done!"

install-config:
	install -d -m 755 $(DESTDIR)$(sysconfdir)/stplr
	[ -f $(DESTDIR)$(sysconfdir)/stplr/stplr.toml ] || install -m 644 $(DEFAULT_CONF) $(DESTDIR)$(sysconfdir)/stplr/stplr.toml

install-sysusers:
	install -Dpm644 $(SYSUSERS_CONF) $(DESTDIR)$(SYSUSERS_DIR)/stplr.conf

install-tmpfiles:
	install -Dpm644 $(TMPFILES_CONF) $(DESTDIR)$(TMPFILES_DIR)/stplr.conf

install-cachedir:
	install -d -m 755 $(DESTDIR)$(CACHE_DIR)

install-post:
	@echo "Running post-installation system setup..."
	setcap cap_setuid,cap_setgid+ep $(DESTDIR)$(bindir)/$(NAME) || echo "Skipping setcap (insufficient permissions?)"
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
		$(DESTDIR)$(datadir)/zsh/site-functions/_$(NAME)

clean clear:
	rm -f $(BIN)

# Development Targets

fmt:
	$(GOLANGCI_LINT_BIN) run --fix

i18n:
	$(XGOTEXT_BIN) --output ./internal/translations/default.pot
	msguniq --use-first -o ./internal/translations/default.pot ./internal/translations/default.pot
	msgmerge --backup=off -U ./internal/translations/po/ru/default.po ./internal/translations/default.pot

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
