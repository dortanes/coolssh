import {
  BoxIcon,
  BriefcaseIcon,
  CloudIcon,
  CodeIcon,
  DatabaseIcon,
  FlaskConicalIcon,
  FolderIcon,
  GlobeIcon,
  HouseIcon,
  type LucideIcon,
  MonitorIcon,
  ServerIcon,
  ShieldIcon,
  SquareTerminalIcon,
} from 'lucide-react';
import { Icon } from '#bindings/library/models.ts';
import { m } from '@/paraglide/messages.js';

type IconChoice = { icon: Icon; Glyph: LucideIcon; label: () => string };

const shared: IconChoice[] = [
  { icon: Icon.ServerIcon, Glyph: ServerIcon, label: m.icon_server },
  { icon: Icon.DatabaseIcon, Glyph: DatabaseIcon, label: m.icon_database },
  { icon: Icon.CloudIcon, Glyph: CloudIcon, label: m.icon_cloud },
  { icon: Icon.GlobeIcon, Glyph: GlobeIcon, label: m.icon_globe },
  { icon: Icon.HomeIcon, Glyph: HouseIcon, label: m.icon_home },
  { icon: Icon.BriefcaseIcon, Glyph: BriefcaseIcon, label: m.icon_briefcase },
  { icon: Icon.CodeIcon, Glyph: CodeIcon, label: m.icon_code },
  { icon: Icon.ShieldIcon, Glyph: ShieldIcon, label: m.icon_shield },
  { icon: Icon.TerminalIcon, Glyph: SquareTerminalIcon, label: m.icon_terminal },
  { icon: Icon.FlaskIcon, Glyph: FlaskConicalIcon, label: m.icon_flask },
  { icon: Icon.BoxIcon, Glyph: BoxIcon, label: m.icon_box },
];

/** The icons a folder can take, each drawn as what it names, in the order the picker shows them. */
export const folderIcons: IconChoice[] = [
  { icon: Icon.FolderIcon, Glyph: FolderIcon, label: m.icon_folder },
  ...shared,
];

const monitor: IconChoice = { icon: Icon.MonitorIcon, Glyph: MonitorIcon, label: m.icon_monitor };

/** The icons a folder can draw its hosts with, the same set as Go's. */
export const hostIcons: IconChoice[] = [monitor, ...shared];

/** The glyph a folder is drawn with; an icon the app does not know falls back to the folder. */
export function folderGlyph(icon: Icon): LucideIcon {
  return folderIcons.find((choice) => choice.icon === icon)?.Glyph ?? FolderIcon;
}

/** The choice a host icon is drawn and named with; an unknown one falls back to the monitor. */
export function hostChoice(icon: Icon): IconChoice {
  return hostIcons.find((choice) => choice.icon === icon) ?? monitor;
}
