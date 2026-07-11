# AGENTS.md

## Toolchain (mise)

This project uses [**mise**](https://mise.jdx.dev) to pin tools, expose tasks, and wire git hooks. `mise.toml` is the source of truth. Don't install tools by hand or add ad-hoc scripts; add a mise tool or task instead.

**Setup** (once, and per new worktree): `mise trust && mise run setup`.

**Run via mise.** Run `mise run check` before you call work done. A few examples, not the full list:

```sh
mise run check          # all linters/formatters/validators (alias: lint); add --fix to auto-fix
mise run test           # go test -race ./...
mise tasks              # discover every task
mise run <task> --help  # a task's flags
```

Prefer `mise run <task>` over calling the tool directly, so local, hooks, and CI stay in sync.

## Git hooks (hk)

Commits run [hk](https://hk.jdx.dev), the same `check` CI runs, to format and lint staged files. Fix failures with `mise run check --fix`. Don't disable steps to push a commit through; `git commit --no-verify` skips hooks for a WIP commit.

## Project notes

- Go library: package root is the module. `vendor/` is committed and excluded from lint via `hk.pkl`'s `commonIgnores`.
- Travis CI was replaced by GitHub Actions (`.github/workflows/check.yml` + `test.yml`).

## Extending the setup

Changing tools, tasks, env, or hooks? Edit the config, don't bolt on scripts, then run `mise run check`. Where things live:

- [`mise.toml`](mise.toml): the source of truth for `[tools]`, `[tasks]`, `[vars]`, `[settings]`, and `[hooks]`.
- [`mise.lock`](mise.lock): resolved versions plus checksums. Commit it; regenerate with `mise install` then `mise lock --platform macos-arm64,linux-x64` after a `[tools]` change.
- [`.mise/`](.mise/): project-local state (gitignored stamp), like the setup stamp the `setup`/`enter` hooks read.
- [`hk.pkl`](hk.pkl): the pre-commit and `check` pipeline (linters and formatters, in Pkl). Add or edit a lint step here.
- Linter config scaffolds live at the repo root ([`typos.toml`](typos.toml), [`.betterleaks.toml`](.betterleaks.toml), [`lychee.toml`](lychee.toml), [`rumdl.toml`](rumdl.toml), [`.yamllint`](.yamllint)).

For tool, task, and hook syntax, see the [mise](https://mise.jdx.dev) and [hk](https://hk.jdx.dev) docs.
