package main

import (
	"image/color"
	server_ssh "swiss-army-shell/server"
	"swiss-army-shell/utils"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var changedDir bool = false
var command string = ""
var consoleOut string = ""
var AppSession server_ssh.ServerSession
var MainWindow fyne.Window

func SettingsButton() fyne.CanvasObject {
	btn := widget.NewButton("Settings", func() {})
	return btn
}

func AboutButton() fyne.CanvasObject {
	btn := widget.NewButton("About", func() {})
	return btn
}

func Margin(size int) fyne.CanvasObject {
	margin := canvas.NewRectangle(color.Transparent)
	margin.Resize(fyne.NewSize(100, float32(size)))
	return margin
}

func AppBar() fyne.CanvasObject { // appbar prototype
	bb := BackButton()
	newCol := color.RGBAModel.Convert(color.NRGBA{R: 60, G: 60, B: 60, A: 60})
	border := canvas.NewRectangle(newCol)
	border.SetMinSize(fyne.NewSize(100, 2))
	margin := canvas.NewRectangle(color.Transparent)
	margin.Resize(fyne.NewSize(100, 10))
	bg := container.New(layout.NewBorderLayout(nil, container.NewVBox(margin, border), nil, nil))
	str := AppSession.Wd
	label := widget.NewLabel(str)
	label.TextStyle = fyne.TextStyle{
		Bold:      true,
		Italic:    false,
		Monospace: false,
		Symbol:    false,
		TabWidth:  0,
	}
	stack := container.NewStack(bg, container.NewHBox(bb, label, layout.NewSpacer(), AboutButton(), SettingsButton()))
	return stack
}

func FileListView(app fyne.App) fyne.CanvasObject {
	var btns []fyne.CanvasObject
	filesList, err := AppSession.ListFiles()
	if err != nil {
		println("error ocurred during file listing")
	}
	for _, f := range filesList {
		if f.IsDir {
			btns = append(btns, FolderButton(f, app))
		} else {
			btns = append(btns, FileButton(f, app))
		}
	}
	return container.NewVScroll(container.NewVBox(btns...))
}

func FileButton(f server_ssh.File, app fyne.App) fyne.CanvasObject { // a file will be passed into this
	fileButtonContent := container.New(layout.NewHBoxLayout(), widget.NewIcon(theme.FileIcon()), widget.NewLabel(f.Name), layout.NewSpacer(), DeleteButton(f), PropertiesButton(f, app))
	fileButton := NewWidgetButton(func() {
		ShowFileEditorWindow(f, app)
	}, fileButtonContent)
	return fileButton
}

func BackButton() fyne.CanvasObject {
	return widget.NewButton("Back", func() {
		if AppSession.ChangeWD("../") != nil {
			println("Error ocurred during WD changing")
		}
		changedDir = true
	})
}

func PropertiesButton(p server_ssh.File, app fyne.App) fyne.CanvasObject {
	propertyClosure := func() {
		pWin := app.NewWindow("Properties")
		pWinCont := container.New(layout.NewVBoxLayout(),
			widget.NewLabel("File name: "+p.Name),
			widget.NewLabel("Size: "+utils.BytesHumanReadable(p.Size)),
			widget.NewLabel("Last modified: "+p.Modified),
			widget.NewLabel("Permissions: "+p.Mode))
		pWin.SetContent(pWinCont)
		pWin.Show()
	}
	return widget.NewButtonWithIcon("", theme.ListIcon(), propertyClosure)
}

func DeleteButton(f server_ssh.File) fyne.CanvasObject {
	return widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		dialog.ShowConfirm("Delete", "Do you wish to delete the file/folder?", func(b bool) {
			if b {
				AppSession.RemoveFileTarget(f.Name)
				MainWindow.Content().Refresh()
			}
		}, MainWindow)
	})
}

func FolderButton(f server_ssh.File, app fyne.App) fyne.CanvasObject {
	folderButtonContent := container.New(
		layout.NewHBoxLayout(),
		widget.NewIcon(theme.FolderIcon()),
		widget.NewLabel(f.Name),
		layout.NewSpacer(),
		DeleteButton(f),
		PropertiesButton(f, app))
	folderButton := NewWidgetButton(func() {
		err := AppSession.ChangeWD(f.Name)
		if err != nil {
			println("Serious problem")
			println(err.Error())
		}
		changedDir = true
	}, folderButtonContent)
	return folderButton
}

func ExecutiveTextBox() fyne.CanvasObject {
	entry := widget.NewEntry()
	entry.OnChanged = func(s string) {
		command = s
	}
	size := fyne.NewSize(800, 38)
	packagedEntry := container.New(layout.NewGridWrapLayout(size), entry)
	return packagedEntry
}

func ExecutionButton(app fyne.App) fyne.CanvasObject {
	button := widget.NewButton("Execute", func() {
		cOut, err := AppSession.ExecuteRaw(command)
		consoleOut = string(cOut)
		if err != nil {
			println("Error occurred during execution")
			println(consoleOut)
		}
		dialog.ShowInformation("Output", consoleOut, MainWindow)
	})
	size := fyne.NewSize(100, 38)
	packagedButton := container.New(layout.NewGridWrapLayout(size), button)
	return packagedButton
}

func ExecutiveShellWidget(app fyne.App) fyne.CanvasObject {
	etb := ExecutiveTextBox()
	eb := ExecutionButton(app)
	hbox := container.NewHBox(etb, eb)
	return container.New(layout.NewCenterLayout(), hbox)
}

func MainShellWindowContent(app fyne.App) fyne.CanvasObject {
	flv := FileListView(app)
	esw := ExecutiveShellWidget(app)
	ab := AppBar()
	mainWinCont := container.New(layout.NewBorderLayout(ab, esw, nil, nil), flv, esw, ab)
	return mainWinCont
}

func ShowFileEditorWindow(f server_ssh.File, app fyne.App) {
	eWindow := app.NewWindow(f.Name)
	eWindow.Resize(fyne.NewSize(500, 500))
	fileEditor := widget.NewEntry()
	bytes, err := AppSession.ReadFileInput(f.Name) // change for changing cwd
	if err == nil {
		fileEditor.Text = string(bytes)
	} else {
		eWindow.SetContent(widget.NewLabel("Failed to open file"))
		eWindow.Show()
	}
	fileTools := container.New(layout.NewHBoxLayout(), widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), func() {
		AppSession.ReWriteFile(f.Name, fileEditor.Text)
	}),
		widget.NewButtonWithIcon("Exit", theme.NavigateBackIcon(), func() {
			eWindow.Close()
		}), widget.NewButtonWithIcon("Save & Exit", theme.DocumentCreateIcon(), func() {
			AppSession.ReWriteFile(f.Name, fileEditor.Text)
			eWindow.Close()
		})) // save button, exit button, save and exit button
	eWindowContent := container.New(layout.NewBorderLayout(fileTools, nil, nil, nil), fileTools, fileEditor)
	eWindow.SetContent(eWindowContent)
	eWindow.Show()
}

func ShellWindow(app fyne.App) fyne.Window {
	window := app.NewWindow("SwissArmyShell")
	MainWindow = window
	mainContStart := func() {
		window.SetContent(MainShellWindowContent(app))
		go func() {
			for range time.Tick(time.Millisecond) {
				if changedDir {
					window.SetContent(MainShellWindowContent(app))
					changedDir = false
				}
			}
		}()
	}
	// first the login...
	var username string
	var password string
	var target string
	usernameEntry := widget.NewEntry()
	usernameEntry.OnChanged = func(val string) {
		username = val
	}
	usernameEntry.PlaceHolder = "Enter username"
	passEntry := widget.NewEntry()
	passEntry.Password = true
	passEntry.OnChanged = func(val string) {
		password = val
	}
	passEntry.PlaceHolder = "Enter password"
	targetEntry := widget.NewEntry()
	targetEntry.PlaceHolder = "Enter server"
	targetEntry.OnChanged = func(val string) {
		target = val
	}
	loginForm := container.New(layout.NewVBoxLayout(), usernameEntry, passEntry, targetEntry, widget.NewButton("Login", func() {
		sesh, err := server_ssh.ConnectSSH(username, password, target, "")
		AppSession = sesh
		if err != nil {
			println("Terrible misfortune")
			dialog.ShowInformation("Login failed", "Failed to log in, check your credentials or if your server is still active", window)
			return
		}
		err1 := sesh.ConnectSFTP()
		if err1 != nil {
			dialog.ShowInformation("Login failed", "Failed to log in, check your credentials or if your server is still active", window)
			return
		} else {
			mainContStart()
		}

	}))
	window.Resize(fyne.NewSize(500, 500))
	window.SetContent(loginForm)
	return window
}

func main() {
	sashApp := app.New()
	window := ShellWindow(sashApp)
	window.ShowAndRun()
}
