package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	xWidget "fyne.io/x/fyne/widget"
)

const CorpusMacroReplacerDefaultPath = `C:\Tri D Corpus\CorpusMacroReplacer\`
const CorpusFileBrowserDefaultPath = `file://C:\Tri D Corpus\Corpus 5.0\elmsav\`

var CorpusFileBrowserFreqentlyUsed = []string{
	`C:\Tri D Corpus\Corpus 4.0\elmsav\`,
	`C:\Tri D Corpus\Corpus 4.0\sobasav\`,
	`C:\Tri D Corpus\Corpus 5.0\elmsav\`,
	`C:\Tri D Corpus\Corpus 5.0\sobasav\`,
	`C:\Tri D Corpus\Corpus 6.0\elmsav\`,
	`C:\Tri D Corpus\Corpus 6.0\sobasav\`,
}

const MacrosDefaultPathNormal = `C:\Tri D Corpus\Corpus 5.0\Makro\`
const MacrosDefaultPath = `file://C:\Tri D Corpus\Corpus 5.0\Makro\`

var loaddedFileForPreview *ElementFile
var loaddedFileForPreviewError error
var SelectedPath string

// var centerViewWithCorpusPreview *fyne.Container

// var macrosToBeChanged []string = []string{}
var loadedFiles []struct {
	path   string
	isFile bool
} = []struct {
	path   string
	isFile bool
}{}
var ListOfLoadedFilesContainer *fyne.Container
var outputPath string

var DialogSizeDefault fyne.Size = fyne.NewSize(950, 650)

func RemoveIndex[T any](s []T, index int) []T {
	return append(s[:index], s[index+1:]...)
}

func refreshListOfLoadedFiles() {
	if ListOfLoadedFilesContainer == nil {
		return
	}
	ListOfLoadedFilesContainer.RemoveAll() // slow...
	listOfLoadedFiles := widget.NewList(
		func() int {
			return len(loadedFiles)
		},
		func() fyne.CanvasObject {
			l := widget.NewLabel("")
			// l.Wrapping = fyne.TextWrap
			l.Truncation = fyne.TextTruncateEllipsis
			l.Alignment = fyne.TextAlignLeading
			return container.NewBorder(nil, nil,
				widget.NewIcon(theme.FileIcon()),
				widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
					refreshCorpusPreviewFunc()
				}),
				l,
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			hbox := o.(*fyne.Container).Objects
			if !loadedFiles[i].isFile {
				icon := hbox[1].(*widget.Icon)
				icon.SetResource(theme.FolderIcon())
			}
			l := hbox[0].(*widget.Label)
			l.SetText(loadedFiles[i].path)
			b := hbox[2].(*widget.Button)
			b.OnTapped = func() {
				loadedFiles = RemoveIndex(loadedFiles, i)
				ListOfLoadedFilesContainer.Refresh()
			}
		})
	ListOfLoadedFilesContainer.Add(listOfLoadedFiles)
}

func CorpusFileTreeOnSelected(uid widget.TreeNodeID) {
	parsedURL, err := url.Parse(string(uid))
	if parsedURL.Scheme != "file" {
		log.Println("URL scheme must be 'file'", uid)
	}
	path := parsedURL.Path
	SelectedPath = parsedURL.Host + path
	stat, err := os.Stat(SelectedPath)
	if err == nil {
		if !stat.IsDir() && isCorpusExtension(SelectedPath) {
			elementFile, err := NewCorpusFile(SelectedPath)
			loaddedFileForPreview = elementFile
			if err != nil {
				log.Println(err)
			}
			loaddedFileForPreviewError = err
		}
	}
	if refreshCorpusPreviewFunc != nil {
		refreshCorpusPreviewFunc()
	}
}

func getLeftPanel(myWindow *fyne.Window) *fyne.Container {
	nothingOpen := widget.NewLabel("Brak otwartych plików!")
	nothingOpen.Truncation = fyne.TextTruncateClip
	nothingOpen.Alignment = fyne.TextAlignCenter

	typical := widget.NewLabel("Typowe ścieżki:")
	typical.Truncation = fyne.TextTruncateClip

	var CorpusFileTreeContainer *fyne.Container
	var fileTree *xWidget.FileTree

	defaultContainer := container.NewVBox(nothingOpen, typical)
	for _, p := range CorpusFileBrowserFreqentlyUsed {
		_, err := os.Stat(p)
		if err == nil {
			defaultContainer.Add(widget.NewButtonWithIcon(p, theme.FolderOpenIcon(), func() {
				// pretty fat finger solution, duplicated below...
				CorpusFileTreeContainer.Remove(fileTree)
				fileTree = xWidget.NewFileTree(storage.NewFileURI(p))
				fileTree.OnSelected = CorpusFileTreeOnSelected
				// todo filter extension but allow directory
				// fileTree.Filter = storage.NewExtensionFileFilter([]string{".S3D", ".E3D"})
				CorpusFileTreeContainer.Remove(defaultContainer)
				CorpusFileTreeContainer.Add(fileTree)
				CorpusFileTreeContainer.Refresh()
			}))
		}
	}
	var openFilesButton *widget.Button = widget.NewButtonWithIcon("Otwórz katalog", theme.FolderOpenIcon(), func() {
		folderDialog := dialog.NewFolderOpen(func(lu fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, *myWindow)
				return
			}
			if lu == nil {
				return
			}
			CorpusFileTreeContainer.Remove(fileTree)
			fileTree = xWidget.NewFileTree(storage.NewFileURI(lu.Path()))
			fileTree.OnSelected = CorpusFileTreeOnSelected
			fileTree.Filter = storage.NewExtensionFileFilter([]string{".S3D", ".E3D"})
			CorpusFileTreeContainer.Remove(defaultContainer)
			CorpusFileTreeContainer.Add(fileTree)
			CorpusFileTreeContainer.Refresh()
		}, *myWindow)

		uri, err := storage.ParseURI(CorpusFileBrowserDefaultPath)
		if err == nil {
			listable, err := storage.ListerForURI(uri)
			if err == nil {
				folderDialog.SetLocation(listable)
			}
		}
		folderDialog.Resize(DialogSizeDefault)
		folderDialog.Show()
	})
	CorpusFileTreeContainer = container.NewBorder(openFilesButton, nil, nil, nil, defaultContainer)

	files := widget.NewAccordionItem("Pliki Corpusa", CorpusFileTreeContainer)
	files.Open = true
	return CorpusFileTreeContainer
}

var MacrosToChangeEntries []*widget.Entry

func getMacroName(path string) string {
	base, found := strings.CutSuffix(path, ".CMK") // todo sla
	if found {
		return filepath.Base(base)
	}
	return filepath.Base(path)
}

func RemoveFromSlice(rem *widget.Entry) {
	c := MacrosToChangeEntries
	for i, o := range c {
		if o != rem {
			continue
		}

		removed := make([]*widget.Entry, len(c)-1)
		copy(removed, c[:i])
		copy(removed[i:], c[i+1:])

		MacrosToChangeEntries = removed
		return
	}
}

var addMakroButton *widget.Button

func getRightPanel(myWindow *fyne.Window) *widget.Accordion {
	macrosToChangeContainer := container.NewVBox()

	addMakroButton = widget.NewButton("Dodaj makro do zamiany", func() {
		newMacroInput := widget.NewEntry()
		newEntryLabel := widget.NewLabelWithStyle("Nic nie wybrano", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		newEntryLabel.Truncation = fyne.TextTruncateEllipsis
		newMacroInput.SetPlaceHolder(`C:\Tri D Corpus\Corpus 5.0\Makro\*.CMK`)
		newMacroInput.OnChanged = func(path string) {
			stat, err := os.Stat(path)
			if err == nil && stat != nil && !stat.IsDir() {
				newEntryLabel.SetText(getMacroName(path))
			} else {
				newEntryLabel.SetText("Plik nie istnieje!")
			}
			newEntryLabel.Refresh()
		}
		MacrosToChangeEntries = append(MacrosToChangeEntries, newMacroInput) // Add to the list
		var row *fyne.Container
		row = container.NewBorder(newEntryLabel, nil, nil,
			container.NewHBox(
				widget.NewButtonWithIcon("", theme.FolderOpenIcon(), func() {
					fileOpenDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
						if err != nil {
							dialog.ShowError(err, *myWindow)
							return
						}
						path := reader.URI().Path()
						newMacroInput.SetText(path)
						newEntryLabel.SetText(getMacroName(path))
						row.Refresh()
					}, *myWindow)

					uri, err := storage.ParseURI(MacrosDefaultPath)
					if err == nil {
						listable, err := storage.ListerForURI(uri)
						if err == nil {
							fileOpenDialog.SetLocation(listable)
						}
					}
					fileOpenDialog.SetFilter(storage.NewExtensionFileFilter([]string{".CMK"}))
					fileOpenDialog.Resize(DialogSizeDefault)
					fileOpenDialog.Show()
				}),
				widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
					macrosToChangeContainer.Remove(row)
					RemoveFromSlice(newMacroInput)
					macrosToChangeContainer.Refresh()
					refreshCorpusPreviewFunc()
				}),
			),
			newMacroInput,
		)
		macrosToChangeContainer.Add(row)
	})
	addMakroButton.OnTapped()

	item1 := widget.NewAccordionItem("Makra do zamiany",
		container.NewVBox(
			addMakroButton,
			macrosToChangeContainer,
		),
	)
	item1.Open = true
	ListOfLoadedFilesContainer = container.NewBorder(nil, nil, nil, nil)
	refreshListOfLoadedFiles()
	item2 := widget.NewAccordionItem("Wybrane pliki/foldery", ListOfLoadedFilesContainer)
	item2.Detail.Refresh()
	item2.Open = true
	accordion := widget.NewAccordion(item1, item2)
	accordion.MultiOpen = true
	return accordion
}

var refreshCorpusPreviewFunc func()
var corpusPreviewContainer *ElementFileContainer

// var corpusPreviewPath string

func isCorpusExtension(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return slices.Contains([]string{".e3d", ".s3d"}, ext)
}

func generateRefreshCorpusPreview(vBox *fyne.Container, toolbarLabel *toolbarLabel) func() {
	refreshCorpusPreviewFunc = func() {
		vBox.RemoveAll()
		stat, err := os.Stat(SelectedPath)
		if err != nil {
			l := widget.NewLabel(fmt.Sprintf("Nie można przeczytać '%s'\n: %s", SelectedPath, err))
			l.Wrapping = fyne.TextWrapWord
			vBox.Add(l)
		} else if stat.IsDir() {
			l := widget.NewLabel(fmt.Sprintf("Wybrano katlog '%s'", SelectedPath))
			l.Wrapping = fyne.TextWrapWord
			l2 := widget.NewLabel("Kliknij Dodaj aby wybrać wszystkie pliki w katalogu")
			l3 := widget.NewLabel("Kliknij w plik aby otworzyć podgląd")
			vBox.Add(l)
			vBox.Add(l2)
			vBox.Add(l3)
		} else {
			if !isCorpusExtension(SelectedPath) {
				l := widget.NewLabel("To nie jest plik Corpusa: ")
				l.Wrapping = fyne.TextWrapWord
				lErr := widget.NewLabel(fmt.Sprintf("%s", SelectedPath))
				lErr.Wrapping = fyne.TextWrapWord
				vBox.Add(l)
				vBox.Add(lErr)
				return
			}
			if loaddedFileForPreviewError != nil {
				l := widget.NewLabel("Błąd podczas czytania Corpusa:")
				l.Wrapping = fyne.TextWrapWord
				lErr := widget.NewLabel(fmt.Sprintf("%s", loaddedFileForPreviewError))
				lErr.Wrapping = fyne.TextWrapWord
				vBox.Add(l)
				vBox.Add(lErr)
				return
			}
			if loaddedFileForPreview == nil {
				return
			}
			o := toolbarLabel.ToolbarObject()
			var container *fyne.Container = o.(*fyne.Container)
			headerLabel := container.Objects[1].(*widget.Label)
			headerLabel.SetText("Podgląd: " + SelectedPath)
			if corpusPreviewContainer == nil || corpusPreviewContainer.elementFile != loaddedFileForPreview {
				elemFileCon := NewElementFileContainer(loaddedFileForPreview)
				corpusPreviewContainer = elemFileCon
				vBox.Add(elemFileCon)
			} else {
				//if corpusPreview.elementFile == loaddedFileForPreview {
				corpusPreviewContainer.elementFile = loaddedFileForPreview
				corpusPreviewContainer.Refresh()
				vBox.Add(corpusPreviewContainer)
				// } else {
				// 	log.Println("something webt wrong when refreshing corpus preview")
			}
		}
	}
	return refreshCorpusPreviewFunc
}

func addToLoadedFilesAndRefresh(path string) {
	contains := slices.ContainsFunc(loadedFiles, func(e struct {
		path   string
		isFile bool
	}) bool {
		return e.path == path
	})
	if path != "" && !contains {
		stat, err := os.Stat(path)
		isFile := true
		if err == nil && stat.IsDir() {
			isFile = false
		}
		loadedFiles = append(loadedFiles, struct {
			path   string
			isFile bool
		}{path, isFile})
		refreshListOfLoadedFiles()
	}
}

func getCenterPanel(w fyne.Window) *fyne.Container {
	vBox := container.NewVBox()
	// centerViewWithCorpusPreview = vBox // setting global reference, not ideal...
	toolbarLabel := NewToolbarLabel("Wybierz plik aby podejrzeć")

	topBar := widget.NewToolbar(
		toolbarLabel,
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(resourceFilterSvg, func() {
			// popUpContent := container.NewVBox(
			// 	widget.NewButton("Close", func() {
			// 	}),
			// )
			// popup := widget.NewModalPopUp(popUpContent, w.Canvas())
			// popup.Show()
			// var filterDialog
			check := widget.NewCheck("Pokazuj tylko elementy z przynajmniej jednym makrem", func(b bool) {
				Settings.hideElementsWithZeroMacros = b
			})
			check.Checked = Settings.hideElementsWithZeroMacros
			filterDialog := dialog.NewCustom("Filtruj", "Ok", container.NewVBox(
				check,
			), w)
			filterDialog.Show()
			filterDialog.SetOnClosed(func() {
				if corpusPreviewContainer == nil {
					return
				}
				corpusPreviewContainer.Update(loaddedFileForPreview)
			})
		}),
		NewToolbarButtonWithIcon("Wczytaj plik/katalog", theme.ContentAddIcon(), func() {
			addToLoadedFilesAndRefresh(SelectedPath)
		},
		),
	)
	generateRefreshCorpusPreview(vBox, toolbarLabel)

	scrollable := container.NewScroll(
		vBox,
	)
	var centerPanel *fyne.Container = container.NewBorder(topBar, nil, nil, nil, scrollable)

	return centerPanel
}

func RunGui() {
	myApp := app.NewWithID("pl.net.stolarz")
	myWindow := myApp.NewWindow("Corpus Makro Replacer")
	myWindow.SetIcon(resourceCorpusreplacerlogoPng)

	centerContainer := getCenterPanel(myWindow)
	logo := canvas.NewImageFromResource(resourceCorpusreplacerlogoPng)
	logo.FillMode = canvas.ImageFillStretch
	logo.SetMinSize(fyne.NewSize(30, 30))

	var outputButton widget.ToolbarItem
	outputButton = NewToolbarButtonWithIcon("Podsumuj i zamień makra", theme.MediaPlayIcon(), onTappedOutputPopup(outputButton, myWindow))
	topBar := widget.NewToolbar(
		widget.NewToolbarSpacer(),
		outputButton,
	)
	left := getLeftPanel(&myWindow)
	right := getRightPanel(&myWindow)
	hSplit := container.NewHSplit(left, centerContainer)
	hSplit.SetOffset(0.2)
	center := container.NewHSplit(hSplit, right)
	center.SetOffset(0.8)
	var border *fyne.Container = container.NewBorder(container.NewBorder(nil, nil, logo, nil, topBar), nil, nil, nil, center)

	myWindow.SetContent(border)
	myWindow.Resize(fyne.NewSize(1000, 700))
	myWindow.ShowAndRun()
}
