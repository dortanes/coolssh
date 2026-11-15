# Contributing

## Setup

Install the tools listed under "Build from source" in the [README](README.md), then:

```bash
go tool wails3 dev
```

This runs the app with the frontend's development server; frontend changes reload the window and Go changes rebuild the app.

## Layout

- `main.go` starts the app.
- `internal/desktop`: the windows, the native menu and the service the windows call.
- `internal/preferences`: the settings file and the keyboard shortcuts.
- `internal/library`: hosts, folders, keys and tags, and the settings hosts inherit from their folders.
- `internal/session`: sessions: dialling through jump hosts, authentication, host keys, the terminal stream.
- `internal/hostsource`: folders that take their hosts from a connected service; `netbird` is the one source so far.
- `internal/secrets`: passwords, passphrases and pasted keys in the system keychain.
- `internal/catalog`: the copy the Go side shows and the language resolution.
- `frontend/`: the windows' interface (React, shadcn/ui, Tailwind), built with Rsbuild.
- `frontend/messages`: the catalogs, one per language.
- `frontend/bindings`: generated from the Go service; never edit it by hand.
- `build/`: icons, platform metadata and the build tasks.

## Before you open a pull request

Run every gate:

```bash
go tool wails3 task check
```

It runs Biome, the TypeScript compiler, the frontend tests and build, the bindings freshness check, the duplicate scan, gofmt, `go vet`, staticcheck, gocognit and the Go tests with the race detector. A pull request with a failing gate is not merged.

- After changing the Go service, regenerate the bindings: `go tool wails3 task common:generate:bindings`.
- After changing `build/appicon.icon`, render the icons on macOS with Xcode 26 or later: `go tool wails3 task icons`.
- Components from shadcn/ui are added with its CLI (`pnpm dlx shadcn@latest add <name>` in `frontend/`) and kept as it writes them.

## Copy and translations

Every string the user reads lives in `frontend/messages`, one flat YAML file per language. English (`en.yaml`) is the source of every key; each other catalog has exactly the same keys, which a Go test checks. To add a language, add its catalog, list it in `frontend/project.inlang/settings.json` and in `internal/catalog`, and add it to the language setting.

## Commits

Commit messages follow [Conventional Commits](https://www.conventionalcommits.org): `feat(settings): add the appearance setting`. One logical change per commit. Pull requests are squash-merged, so the pull request's title is the commit that lands, and a check holds it to the same convention.

## Releases

Releases are automated with [release-please](https://github.com/googleapis/release-please). Every push to `main` updates a release pull request that bumps the version in `build/config.yml` and the platform files listed in `release-please-config.json`, and writes `CHANGELOG.md` from the commits since the last release. Merging it tags a draft release, builds the macOS disk image (signed and notarized), the Windows installer and the Linux packages, attaches them with `SHA256SUMS` and build provenance, and publishes the release. Gemini writes the release notes from the changelog and the commits, following `.github/release-notes-prompt.md`; `.github/release-notes.sh` adds the download table, the checks and the links. Once the release is out, the Homebrew cask in `Casks/coolssh.rb` moves to the new version on `main`. A failed build leaves the release as a draft; run the Release workflow by hand with its tag to build and publish it again.

No version number is edited by hand.

### Release configuration

Repository secrets:

| Secret | Contents |
| --- | --- |
| `MACOS_CERTIFICATE_P12`, `MACOS_CERTIFICATE_PASSWORD` | Developer ID Application certificate with its private key, base64-encoded `.p12`, and its password |
| `APPLE_API_KEY_P8`, `APPLE_API_KEY_ID`, `APPLE_API_ISSUER_ID` | App Store Connect API key for notarization, the `.p8` base64-encoded |
| `GEMINI_API_KEY` | Optional Gemini API key for the release notes; without it the notes carry the changelog |
| `RELEASE_PLEASE_TOKEN` | Optional fine-grained token with contents and pull-request write access, so CI runs on the release pull request |

The repository variable `GEMINI_MODEL` picks another Gemini model than `gemini-3.8-flash`. Files go in base64-encoded, for example `base64 -i certificate.p12 | gh secret set MACOS_CERTIFICATE_P12`. Under **Settings → Actions → General**, allow GitHub Actions to create pull requests.

### Signed builds on your Mac

`go tool wails3 task darwin:package:universal` builds `bin/coolssh.app`, and `build/darwin/release.sh` signs it and writes a notarized disk image holding `CoolSSH.app` to `bin/release`. The script needs the Developer ID Application identity in the login keychain, passed as `IDENTITY` by its hash when the keychain holds it twice, and notarization credentials stored once with `xcrun notarytool store-credentials coolssh` or passed as `NOTARY_ARGS`.
