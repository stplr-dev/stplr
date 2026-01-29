module go.stplr.dev/stplr

go 1.25.5

// support of context via github.com/keegancsmith/rpc
// see https://github.com/hashicorp/go-plugin/issues/284
replace github.com/hashicorp/go-plugin => go.stplr.dev/go-plugin v1.7.1-0.20251005131903-98f00a04f159

// different fixes for ./internal/experimental/xtract
// looks like better write own library in the future
replace golift.io/xtractr => go.stplr.dev/xtractr v0.2.3-0.20260121075838-b3378abf763c

// replace golift.io/xtractr => ../xtractr

require (
	github.com/AlecAivazis/survey/v2 v2.3.7
	github.com/PuerkitoBio/purell v1.2.1
	github.com/alecthomas/chroma/v2 v2.23.1
	github.com/bmatcuk/doublestar/v4 v4.10.0
	github.com/charmbracelet/bubbles v0.21.0
	github.com/charmbracelet/bubbletea v1.3.10
	github.com/charmbracelet/lipgloss v1.1.0
	github.com/charmbracelet/x/term v0.2.2
	github.com/coreos/go-systemd/v22 v22.7.0
	github.com/dave/jennifer v1.7.1
	github.com/go-git/go-billy/v5 v5.7.0
	github.com/go-git/go-git/v5 v5.16.4
	github.com/gobwas/glob v0.2.3
	github.com/goccy/go-yaml v1.19.2
	github.com/gofrs/flock v0.13.0
	github.com/google/cel-go v0.26.1
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/goreleaser/nfpm/v2 v2.44.2
	github.com/hashicorp/go-hclog v1.6.3
	github.com/hashicorp/go-plugin v1.7.0
	github.com/jeandeaual/go-locale v0.0.0-20250612000132-0ef82f21eade
	github.com/keegancsmith/rpc v1.3.0
	github.com/knadh/koanf/parsers/toml/v2 v2.2.0
	github.com/knadh/koanf/providers/confmap v1.0.0
	github.com/knadh/koanf/providers/env v1.1.0
	github.com/knadh/koanf/providers/file v1.2.1
	github.com/knadh/koanf/v2 v2.3.2
	github.com/leonelquinteros/gotext v1.7.2
	github.com/mattn/go-isatty v0.0.20
	github.com/mholt/archiver/v4 v4.0.0-alpha.8
	github.com/mitchellh/mapstructure v1.5.0
	github.com/opencontainers/cgroups v0.0.6
	github.com/opencontainers/runc v1.4.0
	github.com/opencontainers/runtime-spec v1.3.0
	github.com/pelletier/go-toml/v2 v2.2.4
	github.com/spf13/afero v1.15.0
	github.com/stretchr/testify v1.11.1
	github.com/urfave/cli/v3 v3.6.2
	go.alt-gnome.ru/capytest v0.0.4
	go.alt-gnome.ru/capytest/providers/podman v0.0.4
	go.elara.ws/vercmp v0.0.0-20250912200949-2e97859b8794
	go.uber.org/mock v0.6.0
	golang.org/x/crypto v0.47.0
	golang.org/x/exp v0.0.0-20260112195511-716be5621a96
	golang.org/x/sys v0.40.0
	golang.org/x/text v0.33.0
	golift.io/xtractr v0.2.3-0.20260111181756-d6376a2e84ce
	google.golang.org/genproto/googleapis/api v0.0.0-20260114163908-3f89685c29c3
	modernc.org/sqlite v1.44.3
	mvdan.cc/sh/v3 v3.12.1-0.20251005234102-d3ff6f655a6a
	xorm.io/xorm v1.3.11
)

require (
	cel.dev/expr v0.24.0 // indirect
	cyphar.com/go-pathrs v0.2.1 // indirect
	dario.cat/mergo v1.0.2 // indirect
	github.com/AlekSi/pointer v1.2.0 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/Masterminds/sprig/v3 v3.3.0 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/ProtonMail/go-crypto v1.3.0 // indirect
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.0 // indirect
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/blakesmith/ar v0.0.0-20190502131153-809d4375e1fb // indirect
	github.com/bodgit/plumbing v1.3.0 // indirect
	github.com/bodgit/sevenzip v1.6.1 // indirect
	github.com/bodgit/windows v1.0.1 // indirect
	github.com/cavaliergopher/cpio v1.0.1 // indirect
	github.com/cavaliergopher/rpm v1.3.0 // indirect
	github.com/charmbracelet/colorprofile v0.2.3-0.20250311203215-f60798e515dc // indirect
	github.com/charmbracelet/harmonica v0.2.0 // indirect
	github.com/charmbracelet/x/ansi v0.10.1 // indirect
	github.com/charmbracelet/x/cellbuf v0.0.13-0.20250311204145-2c3ea96c31dd // indirect
	github.com/checkpoint-restore/go-criu/v7 v7.2.0 // indirect
	github.com/cilium/ebpf v0.17.3 // indirect
	github.com/cloudflare/circl v1.6.1 // indirect
	github.com/containerd/console v1.0.5 // indirect
	github.com/creack/pty v1.1.24 // indirect
	github.com/cyphar/filepath-securejoin v0.6.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dlclark/regexp2 v1.11.5 // indirect
	github.com/dsnet/compress v0.0.1 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/erikgeiser/coninput v0.0.0-20211004153227-1c3628e74d0f // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/gkampitakis/ciinfo v0.3.3 // indirect
	github.com/gkampitakis/go-snaps v0.5.19 // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/rpmpack v0.7.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/goreleaser/chglog v0.7.4 // indirect
	github.com/goreleaser/fileglob v1.4.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/hashicorp/yamux v0.1.2 // indirect
	github.com/huandu/xstrings v1.5.0 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/kdomanski/iso9660 v0.4.0 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/klauspost/compress v1.18.3 // indirect
	github.com/klauspost/pgzip v1.2.6 // indirect
	github.com/knadh/koanf/maps v0.1.2 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/maruel/natural v1.3.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-localereader v0.0.1 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/sys/capability v0.4.0 // indirect
	github.com/moby/sys/mountinfo v0.7.2 // indirect
	github.com/moby/sys/user v0.4.0 // indirect
	github.com/moby/sys/userns v0.1.0 // indirect
	github.com/mrunalp/fileutils v0.5.1 // indirect
	github.com/muesli/ansi v0.0.0-20230316100256-276c6243b2f6 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/muesli/termenv v0.16.0 // indirect
	github.com/ncruces/go-strftime v1.0.0 // indirect
	github.com/nwaples/rardecode/v2 v2.2.2 // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/opencontainers/selinux v1.13.0 // indirect
	github.com/peterebden/ar v0.0.0-20241106141004-20dc11b778e8 // indirect
	github.com/pierrec/lz4/v4 v4.1.23 // indirect
	github.com/pjbgf/sha1cd v0.3.2 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/seccomp/libseccomp-golang v0.11.1 // indirect
	github.com/sergi/go-diff v1.4.0 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/skeema/knownhosts v1.3.1 // indirect
	github.com/sorairolake/lzip-go v0.3.8 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/sshaman1101/dcompress v0.0.0-20200109162717-50436a6332de // indirect
	github.com/stoewer/go-strcase v1.2.0 // indirect
	github.com/stretchr/objx v0.5.3 // indirect
	github.com/syndtr/goleveldb v1.0.0 // indirect
	github.com/therootcompany/xz v1.0.1 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.2.0 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
	github.com/ulikunitz/xz v0.5.15 // indirect
	github.com/vishvananda/netlink v1.3.1 // indirect
	github.com/vishvananda/netns v0.0.5 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	gitlab.com/digitalxero/go-conventional-commit v1.0.7 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	go4.org v0.0.0-20230225012048-214862532bf5 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/term v0.39.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251222181119-0a764e51fe1b // indirect
	google.golang.org/grpc v1.71.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	modernc.org/libc v1.67.6 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
	xorm.io/builder v0.3.13 // indirect
)
