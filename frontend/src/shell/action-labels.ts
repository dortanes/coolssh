import { Action } from '#bindings/preferences/models.ts';
import { m } from '@/paraglide/messages.js';

/** An action Go defines, without the enum's empty value. */
export type ShortcutAction = Exclude<Action, Action.$zero>;

/** What each action is called wherever its shortcut is shown. */
export const actionLabels: Record<ShortcutAction, () => string> = {
  [Action.CommandPaletteAction]: m.palette_title,
  [Action.NewHostAction]: m.library_new_host,
  [Action.NewFolderAction]: m.library_new_folder,
  [Action.ClosePaneAction]: m.shortcuts_close_pane,
  [Action.NextTabAction]: m.shortcuts_next_tab,
  [Action.PreviousTabAction]: m.shortcuts_previous_tab,
  [Action.SplitRightAction]: m.shortcuts_split_right,
  [Action.SplitDownAction]: m.shortcuts_split_down,
  [Action.NextPaneAction]: m.shortcuts_next_pane,
  [Action.PreviousPaneAction]: m.shortcuts_previous_pane,
};

/** The actions in the order Go lists them. */
export const shortcutActions = Object.values(Action).filter(
  (action): action is ShortcutAction => action !== Action.$zero,
);
