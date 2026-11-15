import {
  type Folder,
  type Host,
  Icon,
  type Key,
  type Library,
  type Settings,
} from '#bindings/library/models.ts';

/** Reads a library snapshot for display: what is where and what refers to what. */
export class LibraryView {
  private readonly library: Library;

  constructor(library: Library) {
    this.library = library;
  }

  folder(id: string): Folder | undefined {
    return this.library.folders?.[id];
  }

  host(id: string): Host | undefined {
    return this.library.hosts?.[id];
  }

  keys(): Key[] {
    return this.library.keys ?? [];
  }

  /** Whether an item belongs to a source: a synced folder, or a host such a folder holds. */
  synced(id: string): boolean {
    const folder = this.folder(id);
    return folder ? folder.source != null : Boolean(this.host(id)?.peer);
  }

  /** The folder a synced host's source keeps it in, and the peer it stands for. */
  peerOf(id: string): { folder: string; peer: string } | undefined {
    const peer = this.host(id)?.peer;
    return peer ? { folder: this.parentOf(id), peer } : undefined;
  }

  hosts(): Host[] {
    return Object.values(this.library.hosts ?? {}).filter((host) => host !== undefined);
  }

  /** The ids directly inside parent, the root when parent is empty. */
  children(parent: string): string[] {
    return (parent === '' ? this.library.root : this.folder(parent)?.children) ?? [];
  }

  /** The folder directly holding id, empty for the root. */
  parentOf(id: string): string {
    const holder = Object.values(this.library.folders ?? {}).find((folder) =>
      folder?.children?.includes(id),
    );
    return holder?.id ?? '';
  }

  /** The icon hosts in a folder are drawn with: the nearest set one from it up, else the monitor. */
  hostIconIn(folder: string): Icon {
    for (let id = folder; id !== ''; id = this.parentOf(id)) {
      const icon = this.folder(id)?.hostIcon;
      if (icon) {
        return icon;
      }
    }
    return Icon.MonitorIcon;
  }

  /** How many hosts and folders lie anywhere inside a folder. */
  countInside(folder: string): { hosts: number; folders: number } {
    let hosts = 0;
    let folders = 0;
    for (const id of this.children(folder)) {
      if (this.folder(id)) {
        const inside = this.countInside(id);
        folders += 1 + inside.folders;
        hosts += inside.hosts;
      } else {
        hosts += 1;
      }
    }
    return { hosts, folders };
  }

  /** How many hosts and folders set this key themselves. */
  keyUsers(key: string): number {
    const settings: Settings[] = [
      ...Object.values(this.library.folders ?? {}).map((folder) => folder?.settings),
      ...this.hosts().map((host) => host.settings),
    ].filter((each) => each !== undefined);
    return settings.filter((each) => each.keyId === key).length;
  }
}
