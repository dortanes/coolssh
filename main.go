// Command coolssh is the SSH client.
package main

import (
	"embed"
	"io/fs"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/Xuanwo/go-locale"

	"github.com/dortanes/coolssh/internal/catalog"
	"github.com/dortanes/coolssh/internal/desktop"
	"github.com/dortanes/coolssh/internal/hostsource"
	"github.com/dortanes/coolssh/internal/hostsource/netbird"
	"github.com/dortanes/coolssh/internal/library"
	"github.com/dortanes/coolssh/internal/preferences"
	"github.com/dortanes/coolssh/internal/product"
	"github.com/dortanes/coolssh/internal/secrets"
	"github.com/dortanes/coolssh/internal/session"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed frontend/messages/*.yaml
var messages embed.FS

//go:embed build/config.yml
var buildConfig []byte

//go:embed build/appicon.png
var icon []byte

func main() {
	about, err := product.Read(buildConfig)
	if err != nil {
		log.Fatal(err)
	}
	about.Icon = icon
	folder, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}
	store, err := preferences.Open(filepath.Join(folder, "coolssh", "preferences.json"))
	if err != nil {
		log.Printf("%v; the defaults apply", err)
	}
	hosts, err := library.Open(filepath.Join(folder, "coolssh", "library.json"), secrets.NewKeyring("coolssh"), connectionDefaults())
	if err != nil {
		log.Printf("%v; the library opens empty and the file is kept", err)
	}
	// The user's known_hosts is only read; keys trusted in the app go to its own file.
	userKnownHosts := ""
	if home, err := os.UserHomeDir(); err == nil {
		userKnownHosts = filepath.Join(home, ".ssh", "known_hosts")
	}
	knownHosts := session.NewKnownHosts(userKnownHosts, filepath.Join(folder, "coolssh", "known_hosts"))
	catalogFiles, err := fs.Sub(messages, "frontend/messages")
	if err != nil {
		log.Fatal(err)
	}
	catalogs, err := catalog.Load(catalogFiles)
	if err != nil {
		log.Fatal(err)
	}
	systemLanguages, err := locale.DetectAll()
	if err != nil {
		log.Printf("system languages: %v; the app follows %s", err, catalog.Source)
	}
	frontend, err := fs.Sub(assets, "frontend/dist")
	if err != nil {
		log.Fatal(err)
	}
	app := desktop.New(desktop.Options{
		Product:         about,
		Preferences:     store,
		Library:         hosts,
		KnownHosts:      knownHosts,
		Sources:         []hostsource.Source{netbird.New()},
		Catalogs:        catalogs,
		SystemLanguages: systemLanguages,
		Assets:          frontend,
	})
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

// connectionDefaults are a host's settings when nothing sets them: port 22, the local user, no jump
// host, a keepalive every 30 seconds.
func connectionDefaults() library.Settings {
	name := ""
	if current, err := user.Current(); err == nil {
		// Windows names the user DOMAIN\name; the server knows the name alone.
		name = current.Username[strings.LastIndex(current.Username, `\`)+1:]
	}
	port, keepalive, none, off := 22, 30, "", false
	return library.Settings{
		User:             &name,
		Port:             &port,
		KeyID:            &none,
		Password:         &off,
		JumpHostID:       &none,
		AgentForwarding:  &off,
		KeepaliveSeconds: &keepalive,
	}
}
