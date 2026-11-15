package desktop

import "github.com/wailsapp/wails/v3/pkg/application"

// menu builds the native menu in the language in effect. Roles keep their native actions; the
// label the catalog gives replaces each role's English one.
func (a *App) menu() *application.Menu {
	menu := application.NewMenu()
	app := menu.AddSubmenu(a.options.Product.Name)
	a.role(app, application.About, "menu_about")
	app.AddSeparator()
	app.Add(a.text("menu_settings")).SetAccelerator("CmdOrCtrl+,").OnClick(func(*application.Context) {
		a.openSettings()
	})
	app.AddSeparator()
	a.role(app, application.Hide, "menu_hide")
	a.role(app, application.HideOthers, "menu_hide_others")
	a.role(app, application.ShowAll, "menu_show_all")
	app.AddSeparator()
	a.role(app, application.Quit, "menu_quit")
	edit := menu.AddSubmenu(a.text("menu_edit"))
	a.role(edit, application.Undo, "menu_undo")
	a.role(edit, application.Redo, "menu_redo")
	edit.AddSeparator()
	a.role(edit, application.Cut, "menu_cut")
	a.role(edit, application.Copy, "menu_copy")
	a.role(edit, application.Paste, "menu_paste")
	a.role(edit, application.SelectAll, "menu_select_all")
	a.role(menu, application.WindowMenu, "menu_window")
	return menu
}

// role adds a role's item to menu under the catalog's label; a role the platform lacks adds
// nothing.
func (a *App) role(menu *application.Menu, role application.Role, key string) {
	menu.AddRole(role)
	if item := menu.FindByRole(role); item != nil {
		item.SetLabel(a.text(key))
	}
}
