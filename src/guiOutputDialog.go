package main

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func NewPopupListOfFiles(files []string) *widget.List {
	return widget.NewList(
		func() int {
			return len(files)
		},
		func() fyne.CanvasObject {
			label := widget.NewLabel("")
			label.Truncation = fyne.TextTruncateEllipsis
			label.Alignment = fyne.TextAlignLeading
			return container.NewBorder(nil, nil,
				widget.NewIcon(theme.FileIcon()),
				nil,
				// widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
				// }),
				label,
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			hbox := o.(*fyne.Container).Objects
			l := hbox[0].(*widget.Label)
			l.SetText(files[i])
			// b := hbox[2].(*widget.Button)
			// b.OnTapped = func() {
			// 	files = RemoveIndex(files, i)
			// 	ListOfLoadedFilesContainer.Refresh()
			// }
		})
}

func onTappedOutputPopup(self widget.ToolbarItem, w fyne.Window) func() {
	var popup *widget.PopUp
	// Define an action when the "Show Pop-up" button is tapped

	return func() {
		foundCorpusFiles := []string{}
		for _, dirOrFile := range loadedFiles {
			foundCorpusFiles = append(foundCorpusFiles, findCorpusFiles(dirOrFile.path)...)
		}
		outputPath := widget.NewEntry()
		outputPath.PlaceHolder = CorpusMacroReplacerDefaultPath
		outputPath.SetText(CorpusMacroReplacerDefaultPath)
		logWindow := widget.NewLabel(`
		o list the files 
- has confirmation button
- contains big text field for text. The text should be selectable.
- capture stdout and put is in big text box after confirmation button is clicked
ChatGPT powiedział:
ChatGPT

Here's how you can create a Fyne-based application in Go to meet your requirements. This program includes a dialog with:

    An input for a path.
    A button to list files in the path.
    A confirmation button.
    A large, selectable text field for displaying the output.

Below is the complete Go code:
Key Features of the Code:

    Path Entry: A text entry field (widget.NewEntry) for users to input a directory path.
    List Files Button: Reads the directory and lists files in the output text area.
    Confirmation Button: Runs a sample command (echo "Hello from stdout!") and captures its output.
    Output Display: Uses a multi-line, read-only text widget to display the results or errors.

How It Works:

    The pathEntry allows users to input a directory path.
    Clicking "List Files" reads the directory and lists file names in the output text area.
    Clicking "Run Command" executes a command and displays the captured stdout in the output.

Let me know if you need further modifications!
Powiedziałeś(-aś):
use dialog.NewConfirm
ChatGPT powiedział:
ChatGPT

The code now uses dialog.NewConfirm to display a confirmation dialog when the "Run Command" button is clicked. Users can confirm or cancel the action before executing the command. Let me know if you need further changes!
Powiedziałeś(-aś):
make popup widget scale with window size
ChatGPT powiedział:
ChatGPT

The dialog.NewConfirm popup now dynamically resizes to scale with the window size, taking up 80% of the width and 50% of the height of the main window. Let me know if further adjustments are needed!
Powiedziałeś(-aś):
make the resize dynamic
ChatGPT powiedział:
ChatGPT

The popup now dynamically resizes with the main window. It updates its size both when shown and in response to key events that may resize the window. Let me know if additional adjustments are needed!

		`)
		content := container.NewVSplit(
			container.NewBorder(
				widget.NewLabel(fmt.Sprintf("Znaleziono pliki: %d", len(foundCorpusFiles))),
				nil, nil, nil,
				NewPopupListOfFiles(foundCorpusFiles),
			),
			container.NewBorder(
				// container.NewHBox(
				container.NewBorder(nil, nil,
					widget.NewLabel("Zapisz wybrane pliki do:"),
					container.NewHBox(
						widget.NewButtonWithIcon("", theme.FolderOpenIcon(), func() {
							fileOpenDialog := dialog.NewFolderOpen(func(lu fyne.ListableURI, err error) {
								if err != nil {
									dialog.ShowError(err, w)
									return
								}
								if lu == nil {
									return
								}
								path := lu.Path()
								outputPath.SetText(path)
							}, w)
							fileOpenDialog.Resize(DialogSize)
							fileOpenDialog.Show()
						}),
						widget.NewButtonWithIcon("Wykonaj", theme.MediaPlayIcon(), func() {
							logWindow.SetText(time.Now().String())
						}),
					),
					outputPath,
				), nil, nil, nil,
				container.NewScroll(logWindow),
				// ),
			),
		)

		// Create a container for the pop-up content, including a "Close" button
		popUpContent := container.NewBorder(nil,
			widget.NewButton("Close", func() {
				popup.Hide() // Function to hide the pop-up
			}), nil, nil,
			content,
		)

		popup = widget.NewModalPopUp(popUpContent, w.Canvas())
		popup.Resize(fyne.NewSize(w.Canvas().Size().Width*0.9, w.Canvas().Size().Height*0.9))
		popup.Show()
	}
}
