#!/bin/sh
# Signs bin/coolssh.app with Developer ID, notarizes it, packs it into a disk image as CoolSSH.app and
# notarizes that: bin/release/coolssh-<version>-macos-universal.dmg, both with their tickets stapled.
# Spotlight finds an app by its bundle's file name, split at case changes, so CoolSSH.app answers "ssh".
# IDENTITY picks the Developer ID Application identity, by name or by hash when the keychain holds it twice.
# NOTARY_ARGS are notarytool's credentials, by default a profile stored with
# `xcrun notarytool store-credentials coolssh`.
set -eu
cd "$(dirname "$0")/../.."
identity="${IDENTITY:-Developer ID Application}"
notary_args="${NOTARY_ARGS:---keychain-profile coolssh}"
version="$(/usr/libexec/PlistBuddy -c 'Print :CFBundleShortVersionString' bin/coolssh.app/Contents/Info.plist)"
out=bin/release
image="$out/image"
app="$image/CoolSSH.app"
dmg="$out/coolssh-$version-macos-universal.dmg"
rm -rf "$out"
mkdir -p "$image"
ditto bin/coolssh.app "$app"
codesign --force --options runtime --timestamp --sign "$identity" "$app"
codesign --verify --deep --strict "$app"

notarize() {
	# shellcheck disable=SC2086 # notary_args holds several arguments.
	xcrun notarytool submit "$1" $notary_args --wait
}

# The app carries its own ticket, so a copy taken out of the disk image opens offline too.
zip="$out/coolssh.zip"
ditto -c -k --keepParent "$app" "$zip"
notarize "$zip"
rm "$zip"
xcrun stapler staple "$app"
ln -s /Applications "$image/Applications"
hdiutil create -volname CoolSSH -srcfolder "$image" -fs HFS+ -format UDZO -ov "$dmg"
codesign --force --timestamp --sign "$identity" "$dmg"
notarize "$dmg"
xcrun stapler staple "$dmg"
spctl --assess --type open --context context:primary-signature --verbose=2 "$dmg"
echo "$dmg"
