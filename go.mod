module go.stplr.dev/stplr

go 1.24.4

// see https://github.com/urfave/cli/pull/2196
replace github.com/urfave/cli/v3 => github.com/Maks1mS/urfave-cli/v3 v3.0.0-20250902060525-3de689e94b58

// support of context via github.com/keegancsmith/rpc
// see https://github.com/hashicorp/go-plugin/issues/284
replace github.com/hashicorp/go-plugin => go.stplr.dev/go-plugin v1.7.1-0.20251005131903-98f00a04f159

require (
	gitea.plemya-x.ru/Plemya-x/fakeroot v0.0.2-0.20250408104831-427aaa7713c3
	github.com/AlecAivazis/survey/v2 v2.3.7
	github.com/PuerkitoBio/purell v1.2.1
	github.com/alecthomas/chroma/v2 v2.20.0
	github.com/bmatcuk/doublestar/v4 v4.9.1
	github.com/charmbracelet/bubbles v0.21.0
	github.com/charmbracelet/bubbletea v1.3.10
	github.com/charmbracelet/lipgloss v1.1.0
	github.com/charmbracelet/log v0.4.2
	github.com/charmbracelet/x/term v0.2.2
	github.com/dave/jennifer v1.7.1
	github.com/go-git/go-billy/v5 v5.6.2
	github.com/go-git/go-git/v5 v5.16.3
	github.com/goccy/go-yaml v1.18.0
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/goreleaser/nfpm/v2 v2.43.1
	github.com/hashicorp/go-hclog v1.6.3
	github.com/hashicorp/go-plugin v1.7.0
	github.com/jeandeaual/go-locale v0.0.0-20250612000132-0ef82f21eade
	github.com/keegancsmith/rpc v1.3.0
	github.com/knadh/koanf/parsers/toml/v2 v2.2.0
	github.com/knadh/koanf/providers/confmap v1.0.0
	github.com/knadh/koanf/providers/env v1.1.0
	github.com/knadh/koanf/providers/file v1.2.0
	github.com/knadh/koanf/v2 v2.3.0
	github.com/leonelquinteros/gotext v1.7.2
	github.com/mattn/go-isatty v0.0.20
	github.com/mholt/archiver/v4 v4.0.0-alpha.8
	github.com/mitchellh/mapstructure v1.5.0
	github.com/pelletier/go-toml/v2 v2.2.4
	github.com/spf13/afero v1.15.0
	github.com/stretchr/testify v1.11.1
	github.com/urfave/cli/v3 v3.5.0
	github.com/vmihailenco/msgpack/v5 v5.4.1
	go.alt-gnome.ru/capytest v0.0.3
	go.alt-gnome.ru/capytest/providers/podman v0.0.3
	go.elara.ws/vercmp v0.0.0-20250912200949-2e97859b8794
	go.uber.org/mock v0.6.0
	golang.org/x/crypto v0.43.0
	golang.org/x/exp v0.0.0-20251023183803-a4bb9ffd2546
	golang.org/x/sys v0.38.0
	golang.org/x/text v0.30.0
	modernc.org/sqlite v1.40.0
	mvdan.cc/sh/v3 v3.12.1-0.20251005234102-d3ff6f655a6a
	xorm.io/xorm v1.3.11
)

require (
	dario.cat/mergo v1.0.2 // indirect
	github.com/AlekSi/pointer v1.2.0 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/Masterminds/sprig/v3 v3.3.0 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/ProtonMail/go-crypto v1.3.0 // indirect
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/blakesmith/ar v0.0.0-20190502131153-809d4375e1fb // indirect
	github.com/bodgit/plumbing v1.2.0 // indirect
	github.com/bodgit/sevenzip v1.3.0 // indirect
	github.com/bodgit/windows v1.0.0 // indirect
	github.com/cavaliergopher/cpio v1.0.1 // indirect
	github.com/charmbracelet/colorprofile v0.2.3-0.20250311203215-f60798e515dc // indirect
	github.com/charmbracelet/harmonica v0.2.0 // indirect
	github.com/charmbracelet/x/ansi v0.10.1 // indirect
	github.com/charmbracelet/x/cellbuf v0.0.13-0.20250311204145-2c3ea96c31dd // indirect
	github.com/cloudflare/circl v1.6.1 // indirect
	github.com/connesc/cipherio v0.2.1 // indirect
	github.com/creack/pty v1.1.24 // indirect
	github.com/cyphar/filepath-securejoin v0.4.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dlclark/regexp2 v1.11.5 // indirect
	github.com/dsnet/compress v0.0.1 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/erikgeiser/coninput v0.0.0-20211004153227-1c3628e74d0f // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/gkampitakis/ciinfo v0.3.2 // indirect
	github.com/gkampitakis/go-diff v1.3.2 // indirect
	github.com/gkampitakis/go-snaps v0.5.13 // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/rpmpack v0.7.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/goreleaser/chglog v0.7.3 // indirect
	github.com/goreleaser/fileglob v1.3.0 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/yamux v0.1.2 // indirect
	github.com/huandu/xstrings v1.5.0 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/klauspost/pgzip v1.2.6 // indirect
	github.com/knadh/koanf/maps v0.1.2 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/maruel/natural v1.1.1 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-localereader v0.0.1 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/muesli/ansi v0.0.0-20230316100256-276c6243b2f6 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/muesli/termenv v0.16.0 // indirect
	github.com/ncruces/go-strftime v0.1.9 // indirect
	github.com/nwaples/rardecode/v2 v2.0.0-beta.2 // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/pjbgf/sha1cd v0.3.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/sergi/go-diff v1.3.2-0.20230802210424-5b0b94c5c0d3 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/skeema/knownhosts v1.3.1 // indirect
	github.com/spf13/cast v1.7.1 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/syndtr/goleveldb v1.0.0 // indirect
	github.com/therootcompany/xz v1.0.1 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
	github.com/ulikunitz/xz v0.5.15 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	gitlab.com/digitalxero/go-conventional-commit v1.0.7 // indirect
	go4.org v0.0.0-20200411211856-f5505b9728dd // indirect
	golang.org/x/net v0.45.0 // indirect
	golang.org/x/term v0.36.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241223144023-3abc09e42ca8 // indirect
	google.golang.org/grpc v1.67.3 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	modernc.org/libc v1.66.10 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
	xorm.io/builder v0.3.13 // indirect
)
