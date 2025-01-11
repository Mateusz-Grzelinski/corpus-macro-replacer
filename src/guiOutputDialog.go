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
		logWindow := widget.NewLabel(`todo`)
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
							fileOpenDialog.Resize(DialogSizeDefault)
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
