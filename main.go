package main

import (
	"image/color"
	"io/fs"
	"os"
	"os/exec"
	"strings"
	server_ssh "swiss-army-shell/server"
	"swiss-army-shell/shell"
	"swiss-army-shell/utils"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var changedDir bool = false
var command string = ""
var consoleOut string = ""
var session server_ssh.ServerSession

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
	str, _ := os.Getwd()
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
	filesList, _ := shell.ListFiles()
	for _, f := range filesList {
		if f.IsDir() {
			btns = append(btns, FolderButton(f, app))
		} else {
			btns = append(btns, FileButton(f, app))
		}
	}
	return container.NewVScroll(container.NewVBox(btns...))
}

func FileButton(f os.DirEntry, app fyne.App) fyne.CanvasObject { // a file will be passed into this
	fInfo, _ := f.Info()
	fileButtonContent := container.New(layout.NewHBoxLayout(), widget.NewIcon(theme.FileIcon()), widget.NewLabel(f.Name()), layout.NewSpacer(), PropertiesButton(fInfo, app))
	fileButton := NewWidgetButton(func() {
		ShowFileEditorWindow(f, app)
	}, fileButtonContent)
	return fileButton
}

func BackButton() fyne.CanvasObject {
	return widget.NewButton("Back", func() {
		os.Chdir("../")
		changedDir = true
	})
}

func PropertiesButton(p fs.FileInfo, app fyne.App) fyne.CanvasObject {
	propertyClosure := func() {
		pWin := app.NewWindow("Properties")
		pWinCont := container.New(layout.NewVBoxLayout(),
			widget.NewLabel("File name: "+p.Name()),
			widget.NewLabel("Size: "+utils.BytesHumanReadable(p.Size())),
			widget.NewLabel("Last modified: "+p.ModTime().String()),
			widget.NewLabel("Permissions: "+p.Mode().String()))
		pWin.SetContent(pWinCont)
		pWin.Show()
	}
	return widget.NewButtonWithIcon("", theme.ListIcon(), propertyClosure)
}

func FolderButton(f os.DirEntry, app fyne.App) fyne.CanvasObject {
	fInfo, _ := f.Info()
	folderButtonContent := container.New(
		layout.NewHBoxLayout(),
		widget.NewIcon(theme.FolderIcon()),
		widget.NewLabel(f.Name()),
		layout.NewSpacer(),
		PropertiesButton(fInfo, app))
	folderButton := NewWidgetButton(func() {
		os.Chdir(f.Name())
		changedDir = true
	}, folderButtonContent)
	return folderButton
}

func ExecutiveTextBox() fyne.CanvasObject {
	entry := widget.NewEntry()
	entry.OnChanged = func(s string) {
		command = s
	}
	size := fyne.NewSize(1500, 38)
	packagedEntry := container.New(layout.NewGridWrapLayout(size), entry)
	return packagedEntry
}

func ExecutionButton() fyne.CanvasObject {
	button := widget.NewButton("Execute", func() {
		argv := strings.Split(command, " ")
		cOut, err := exec.Command(argv[0], argv[1:]...).Output()
		if err == nil {
			consoleOut = string(cOut)
		} else {
			consoleOut = "Error"
		}
		println(consoleOut)
	})
	size := fyne.NewSize(100, 38)
	packagedButton := container.New(layout.NewGridWrapLayout(size), button)
	return packagedButton
}

func ExecutiveShellWidget() fyne.CanvasObject {
	etb := ExecutiveTextBox()
	eb := ExecutionButton()
	hbox := container.NewHBox(etb, eb)
	return container.New(layout.NewCenterLayout(), hbox)
}

func MainShellWindowContent(app fyne.App) fyne.CanvasObject {
	flv := FileListView(app)
	esw := ExecutiveShellWidget()
	ab := AppBar()
	mainWinCont := container.New(layout.NewBorderLayout(ab, esw, nil, nil), flv, esw, ab)
	return mainWinCont
}

func ShowFileEditorWindow(f os.DirEntry, app fyne.App) {
	eWindow := app.NewWindow(f.Name())
	fileEditor := widget.NewEntry()
	bytes, err := os.ReadFile(f.Name()) // change for changing cwd
	if err == nil {
		fileEditor.Text = string(bytes)
	} else {
		eWindow.SetContent(widget.NewLabel("Failed to open file"))
		eWindow.Show()
	}
	fileTools := container.New(layout.NewHBoxLayout(), widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), func() {
		os.WriteFile(f.Name(), []byte(fileEditor.Text), os.ModeCharDevice)
	}),
		widget.NewButtonWithIcon("Exit", theme.NavigateBackIcon(), func() {
			eWindow.Close()
		}), widget.NewButtonWithIcon("Save & Exit", theme.DocumentCreateIcon(), func() {
			os.WriteFile(f.Name(), []byte(fileEditor.Text), os.ModeCharDevice)
			eWindow.Close()
		})) // save button, exit button, save and exit button

	eWindowContent := container.New(layout.NewBorderLayout(fileTools, nil, nil, nil), fileTools, fileEditor)
	eWindow.SetContent(eWindowContent)
	eWindow.Show()
}

func ShellWindow(app fyne.App) fyne.Window {
	window := app.NewWindow("SwissArmyShell")
	window.Resize(fyne.NewSize(1000, 1000))
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
	loginForm := container.New(layout.NewVBoxLayout(), widget.NewEntry(), widget.NewEntry(), widget.NewButton("Login", func() {
		mainContStart()
	}))
	window.SetContent(loginForm)
	return window
}

func main() {
	sashApp := app.New()
	window := ShellWindow(sashApp)
	window.ShowAndRun()
}
