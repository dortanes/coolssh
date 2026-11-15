# Roadmap

What CoolSSH does today, what comes next, and what is planned for later. Items in each section are in the order they will be built.

## Done

- Hosts in nested folders that pass connection settings down: user, port, key, password, jump hosts, agent forwarding, keepalive
- Keys picked from a file or pasted, passwords and pasted keys kept in the system keychain
- Sessions in tabs: jump hosts, host key checks, password and keyboard-interactive prompts, reconnect
- Terminal settings: font size, cursor, copy on select, Option as Meta, scrollback, hardware acceleration
- Default user, port, agent forwarding and keepalive for every host
- Folders that stay in sync with a connected service (NetBird so far), filtered by group, with each host's status
- Folder icons, and an icon for the hosts in a folder
- Command palette (⌘K, Ctrl+Shift+K elsewhere): find and connect a host, connect to `user@host` without saving it, run an action; keyboard shortcuts set in the settings
- Split panes: several sessions side by side or stacked in one tab
- Keys: generate a key, copy its public key, install it on a host; see and forget trusted host keys
- Tags on hosts, searchable from the command palette

## Next

1. Import the hosts in the SSH config and keep them in sync

## Later

- File transfer (SFTP): browse a host's files, drag files between the computer and the host
- Port forwarding: saved local, remote and SOCKS forwards that start with a host
- Snippets: saved commands to run in a session or when a host connects
- Typing into several sessions at once
- A local shell in a tab
- Per-host environment variables and a startup command
- Session logs written to a file
- Hardware security keys (FIDO2) and SSH certificates
- Mosh, Telnet and serial connections

## Not planned

- An account or a cloud service of our own. Syncing between computers, if it comes, will use a file you keep in a folder you already sync.
