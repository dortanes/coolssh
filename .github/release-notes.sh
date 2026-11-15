#!/usr/bin/env bash
# Prints the notes of release TAG for the files in ASSETS. Gemini writes the summary, what's new and the
# fixes from release-please's changelog and the commits since the previous published release; without
# GEMINI_API_KEY, or when the request fails, the changelog stands in. The download table, the checks and
# the links are written here, so the model never names a file or a URL.
# Needs TAG, ASSETS, GITHUB_REPOSITORY and GH_TOKEN; GEMINI_MODEL picks the model.
set -euo pipefail
cd "$(dirname "$0")/.."

version="${TAG#v}"
repository="https://github.com/$GITHUB_REPOSITORY"
download="$repository/releases/download/$TAG"
model="${GEMINI_MODEL:-gemini-3.8-flash}"

previous="$(gh release list --repo "$GITHUB_REPOSITORY" --exclude-drafts --exclude-pre-releases --limit 1 --json tagName --jq '.[0].tagName // ""')"
changelog="$(gh release view "$TAG" --repo "$GITHUB_REPOSITORY" --json body --jq .body)"
commits="$(git log --format='- %s%n%w(0,2,2)%b' "${previous:+$previous..}$TAG")"

summary() {
	if [ -z "${GEMINI_API_KEY:-}" ]; then
		echo "::notice::GEMINI_API_KEY is not set; the release notes use the changelog" >&2
		return 1
	fi
	local request response
	request="$(jq -n \
		--rawfile instruction .github/release-notes-prompt.md \
		--arg input "$(printf 'Version: %s\nPrevious release: %s\n\nChangelog:\n%s\n\nCommits:\n%s\n' "$version" "${previous:-none, this is the first release}" "$changelog" "$commits")" \
		'{system_instruction: {parts: [{text: $instruction}]}, contents: [{parts: [{text: $input}]}]}')"
	if ! response="$(curl --silent --show-error --fail-with-body --max-time 120 --retry 2 \
		"https://generativelanguage.googleapis.com/v1beta/models/$model:generateContent" \
		-H "x-goog-api-key: $GEMINI_API_KEY" -H 'Content-Type: application/json' -d "$request")"; then
		echo "::warning::Gemini request failed; the release notes use the changelog: $response" >&2
		return 1
	fi
	jq -er '[.candidates[0].content.parts[] | select(.thought != true) | .text] | join("") | select(test("\\S"))' <<<"$response" ||
		{ echo "::warning::Gemini returned no text; the release notes use the changelog" >&2; return 1; }
}

# Links every file in ASSETS whose name matches one of the patterns, separated by commas.
links() {
	local pattern file name separator=""
	for pattern in "$@"; do
		for file in "$ASSETS"/$pattern; do
			[ -e "$file" ] || continue
			name="${file##*/}"
			printf '%s[%s](%s/%s)' "$separator" "$name" "$download" "$name"
			separator=", "
		done
	done
}

summary || sed -e '1{/^## /d;}' -e '/./,$!d' <<<"$changelog"

cat <<EOF

## Download

| Platform | File | Requirements |
| --- | --- | --- |
| macOS | $(links '*.dmg') | macOS 13 or later, Apple silicon or Intel |
| Windows | $(links '*.exe') | Windows 10 or later, 64-bit |
| Linux | $(links '*.AppImage' '*.deb' '*.rpm' '*.pkg.tar.zst') | GTK 4 and WebKitGTK 6.0, as on Ubuntu 24.04 and Debian 13 |

On macOS, you can also install with Homebrew:

\`\`\`sh
brew tap dortanes/coolssh $repository
brew install --cask dortanes/coolssh/coolssh
\`\`\`

The Windows installer is not code-signed yet, so SmartScreen asks first: choose **More info**, then **Run anyway**.

## Verify your download

Every file is listed in \`SHA256SUMS\` and carries a build provenance attestation from this repository's release workflow:

\`\`\`sh
shasum -a 256 -c SHA256SUMS --ignore-missing
gh attestation verify coolssh-$version-macos-universal.dmg --repo $GITHUB_REPOSITORY
\`\`\`
EOF

if [ -n "$previous" ]; then
	printf '\n**Full changelog:** [%s...%s](%s/compare/%s...%s)\n' "$previous" "$TAG" "$repository" "$previous" "$TAG"
fi
printf '\nFound a bug? [Open an issue](%s/issues/new/choose). Found a vulnerability? Report it privately as described in [SECURITY.md](%s/blob/main/SECURITY.md).\n' "$repository" "$repository"
