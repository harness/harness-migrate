module github.com/harness/harness-migrate/plugin

go 1.26.7

require (
	github.com/drone/go-scm v1.38.9
	github.com/google/uuid v1.6.0
	github.com/harness/cli/v3 v3.7.0
	github.com/harness/harness-migrate v0.0.0
	github.com/spf13/cobra v1.10.2
)

require (
	charm.land/bubbletea/v2 v2.0.9 // indirect
	charm.land/lipgloss/v2 v2.0.6 // indirect
	dario.cat/mergo v1.0.2 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/ProtonMail/go-crypto v1.3.0 // indirect
	github.com/bmatcuk/doublestar v1.3.4 // indirect
	github.com/buildkite/yaml v2.1.0+incompatible // indirect
	github.com/charmbracelet/colorprofile v0.4.3 // indirect
	github.com/charmbracelet/ultraviolet v0.0.0-20260811164956-006e29f97886 // indirect
	github.com/charmbracelet/x/ansi v0.11.8 // indirect
	github.com/charmbracelet/x/term v0.2.2 // indirect
	github.com/charmbracelet/x/termios v0.1.1 // indirect
	github.com/charmbracelet/x/windows v0.2.2 // indirect
	github.com/clipperhouse/displaywidth v0.11.0 // indirect
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	github.com/cloudflare/circl v1.6.1 // indirect
	github.com/cyphar/filepath-securejoin v0.4.1 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/drone/go-convert v0.0.0-20240821195621-c6d7be7727ec // indirect
	github.com/drone/spec v0.0.0-20230920145636-3827abdce961 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/expr-lang/expr v1.17.8 // indirect
	github.com/fatih/color v1.17.0 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-git/gcfg/v2 v2.0.2 // indirect
	github.com/go-git/go-billy/v6 v6.0.0-20250627091229-31e2a16eef30 // indirect
	github.com/go-git/go-git/v6 v6.0.0-20250728093604-6aaf1933ecab // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/gotidy/ptr v1.4.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jedib0t/go-pretty/v6 v6.8.3 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/lmittmann/tint v1.2.0 // indirect
	github.com/lucasb-eyer/go-colorful v1.4.1 // indirect
	github.com/matoous/go-nanoid v1.5.1 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.27 // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/pjbgf/sha1cd v0.4.0 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/schollz/progressbar/v3 v3.13.0 // indirect
	github.com/sergi/go-diff v1.4.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/crypto v0.40.0 // indirect
	golang.org/x/exp v0.0.0-20251219203646-944ab1f22d93 // indirect
	golang.org/x/mod v0.40.0 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/term v0.45.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/harness/harness-migrate => ../

// Replace directives are only honored in the main module, so the parent
// module's go-git pin has to be repeated here.
replace github.com/go-git/go-git/v6 => github.com/go-git/go-git/v6 v6.0.0-20250728093604-6aaf1933ecab
