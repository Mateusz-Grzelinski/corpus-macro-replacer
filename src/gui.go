package main

import (
	"cmp"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"
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
var selectedPath string
var centerViewWithCorpusPreview *fyne.Container

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

var DialogSize fyne.Size = fyne.NewSize(900, 600)

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
				widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {}),
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
	selectedPath = parsedURL.Host + path
	stat, err := os.Stat(selectedPath)
	if err == nil {
		if !stat.IsDir() {
			elementFile, err := ReadCorpusFile(path)
			loaddedFileForPreview = elementFile
			if err != nil {
				log.Println(err)
			}
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
				fileTree.Filter = storage.NewExtensionFileFilter([]string{".S3D", ".E3D"})
				CorpusFileTreeContainer.Remove(defaultContainer)
				CorpusFileTreeContainer.Add(fileTree)
				CorpusFileTreeContainer.Refresh()
			}))
		}
	}
	// CorpusFileTreeContainer = container.NewBorder(nil, nil, nil, nil, defaultContainer)
	// nothingOpen := xWidget.NewFileTree(storage.NewFileURI(CorpusFileBrowserDefaultPath2))
	// nothingOpen.OnSelected = CorpusFileTreeOnSelected
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
		folderDialog.Resize(DialogSize)
		folderDialog.Show()
	})
	CorpusFileTreeContainer = container.NewBorder(openFilesButton, nil, nil, nil, defaultContainer)

	files := widget.NewAccordionItem("Pliki Corpusa", CorpusFileTreeContainer)
	files.Open = true
	return CorpusFileTreeContainer
}

var MacrosToChange []*widget.Entry

func getMacroName(path string) string {
	base, found := strings.CutSuffix(path, ".CMK") // todo sla
	if found {
		return filepath.Base(base)
	}
	return filepath.Base(path)
}

var addMakroButton *widget.Button

func getRightPanel(myWindow *fyne.Window) *widget.Accordion {
	macrosToChangeContainer := container.NewVBox()

	addMakroButton = widget.NewButton("Wybierz makro", func() {
		newMacroInput := widget.NewEntry()
		newEntryLabel := widget.NewLabelWithStyle("Nic nie wybrano", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		newEntryLabel.Truncation = fyne.TextTruncateEllipsis
		newMacroInput.SetPlaceHolder(`C:\Tri D Corpus\Corpus 5.0\Makro\*.CMK`)
		newMacroInput.OnChanged = func(path string) {
			// parsedUri, err := storage.ParseURI(path)
			// canRead, err := storage.CanRead(parsedUri)
			stat, err := os.Stat(path)
			if err == nil && stat != nil && !stat.IsDir() {
				newEntryLabel.SetText(getMacroName(path))
			} else {
				newEntryLabel.SetText("Plik nie istnieje!")
			}
			newEntryLabel.Refresh()
		}
		MacrosToChange = append(MacrosToChange, newMacroInput) // Add to the list
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
					fileOpenDialog.Resize(DialogSize)
					fileOpenDialog.Show()
				}),
				widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
					macrosToChangeContainer.Remove(row)
					macrosToChangeContainer.Refresh()
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

var macroIcon *canvas.Image = canvas.NewImageFromResource(resourceMacroSvg)

func AddMacro(con *fyne.Container, oldMakro *M1) {
	macroIcon.SetMinSize(fyne.NewSquareSize(20))
	con.Add(container.NewHBox(
		// widget.NewCheck("", func(checked bool) {
		// 	return
		// }),
		// widget.NewIcon(theme.MenuDropDownIcon()),
		macroIcon,
		widget.NewRichTextFromMarkdown(fmt.Sprintf("### %s", oldMakro.MakroName)),
		// widget.NewIcon(theme.VisibilityIcon()),
		// widget.NewIcon(theme.DocumentSaveIcon()),
		widget.NewButtonWithIcon("Uaktualnij", theme.NavigateNextIcon(), func() {
			macroGuessedPath := filepath.Join(MacrosDefaultPathNormal, oldMakro.MakroName) + ".CMK"
			for _, makroTochangeName := range MacrosToChange {
				if makroTochangeName.Text == macroGuessedPath {
					return
				}
			}
			needCreation := true
			for _, makroTochangeName := range MacrosToChange {
				if makroTochangeName.Text == "" {
					makroTochangeName.SetText(macroGuessedPath)
					needCreation = false
					break
				}
			}
			if needCreation {
				addMakroButton.OnTapped()
			}
			MacrosToChange[len(MacrosToChange)-1].SetText(macroGuessedPath)
			if refreshCorpusPreviewFunc != nil {
				refreshCorpusPreviewFunc()
			}
		}),
	))

	var newMakro *M1
	for _, makroTochangeName := range MacrosToChange {
		newMakroName := getMacroName(makroTochangeName.Text)
		if oldMakro.MakroName == newMakroName {
			makro, err := LoadMakroFromCMKFile(makroTochangeName.Text)
			newMakro = makro
			if err != nil {
				log.Println(err)
			}
			break
		}
	} // C:\Tri D Corpus\Corpus 5.0\Makro\Zawieszki Standard Stolarz do HDF.CMK

	if newMakro == nil {
		{
			textOld := strings.Join(decodeAllCMKLines(oldMakro.Varijable.DAT), "\n")
			multilineOld := widget.NewRichTextWithText(textOld)
			multilineOld.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
			con.Add(widget.NewRichTextFromMarkdown("`[VARIJABLE]`"))
			con.Add(multilineOld)
		}
		{
			con.Add(widget.NewRichTextFromMarkdown("`[JOINT]`"))
			var textOld string
			if oldMakro.Joint != nil {
				textOld = strings.Join(decodeAllCMKLines(cmp.Or(oldMakro.Joint.DAT, "")), "\n")
			} else {
				textOld = ""
			}
			multilineOld := widget.NewRichTextWithText(textOld)
			multilineOld.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
			con.Add(multilineOld)
		}
	} else {
		{
			textOld := strings.Join(decodeAllCMKLines(oldMakro.Varijable.DAT), "\n")
			multilineOld := widget.NewRichTextWithText(textOld)
			multilineOld.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
			multilineNew := widget.NewMultiLineEntry()
			textNew := strings.Join(decodeAllCMKLines(newMakro.Varijable.DAT), "\n")
			multilineNew.SetText(textNew)
			con.Add(widget.NewRichTextFromMarkdown("`[VARIJABLE]`"))
			con.Add(container.NewHSplit(multilineOld, multilineNew))
		}
		{
			var textOld string
			if oldMakro.Joint != nil {
				textOld = strings.Join(decodeAllCMKLines(oldMakro.Joint.DAT), "\n")
			} else {
				textOld = ""
			}
			multilineOld := widget.NewRichTextWithText(textOld)
			multilineOld.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
			multilineNew := widget.NewMultiLineEntry()
			var textNew string
			if newMakro.Joint != nil {
				textNew = strings.Join(decodeAllCMKLines(newMakro.Joint.DAT), "\n")
			} else {
				textNew = ""
			}
			multilineNew.SetText(textNew)
			con.Add(widget.NewRichTextFromMarkdown("`[JOINT]`"))
			con.Add(container.NewHSplit(multilineOld, multilineNew))
		}
	}
	con.Add(widget.NewRichTextFromMarkdown("`[FORMULE]`\n\nTodo: add button to show the rest"))
}

// // formatka/Daske
func AddPlate(con *fyne.Container, element *Element, adIndex int) {
	vBox := container.NewVBox()
	daskeName := element.Daske.AD[adIndex].DName.Value
	numMakros := 0
	for _, spoj := range element.Elinks.Spoj {
		_adIndex, _ := strconv.Atoi(spoj.O1.Value)
		if _adIndex == adIndex {
			numMakros++
		}
	}
	item1 := widget.NewAccordionItem("▧ Formatka: '"+daskeName+"' (makra: "+strconv.Itoa(numMakros)+")", vBox)
	accordion := widget.NewAccordion(item1)
	con.Add(accordion)
	for _, spoj := range element.Elinks.Spoj {
		_adIndex, _ := strconv.Atoi(spoj.O1.Value)
		if _adIndex != adIndex {
			continue
		}
		AddMacro(vBox, &spoj.Makro1)
	}
}

// todo make refresh instead add
func AddElement(con *fyne.Container, element *Element) {
	cabinetIcon := canvas.NewImageFromResource(resourceCabinetSvg)
	cabinetIcon.SetMinSize(fyne.NewSquareSize(25))
	con.Add(
		container.NewHBox(
			cabinetIcon,
			widget.NewRichTextFromMarkdown(
				fmt.Sprintf(
					"## Szafka: %s (Formatek: %s, Makr: %s)", element.EName.Value, element.Daske.DCount.Value, element.Elinks.COUNT.Value,
				),
			),
		),
	)
	for adIndex, _ := range element.Daske.AD {
		AddPlate(con, element, adIndex)
	}
	con.Add(widget.NewSeparator())
}

var refreshCorpusPreviewFunc func()

func generateRefreshCorpusPreview(vBox *fyne.Container, toolbarLabel *toolbarLabel) func() {
	refreshCorpusPreviewFunc = func() {
		vBox.RemoveAll()
		stat, err := os.Stat(selectedPath)
		if err != nil {
			l := widget.NewLabel(fmt.Sprintf("Nie można przeczytać '%s'\n: %s", selectedPath, err))
			l.Wrapping = fyne.TextWrapWord
			vBox.Add(l)
		} else if stat.IsDir() {
			l := widget.NewLabel(fmt.Sprintf("Wybrano katlog '%s'", selectedPath))
			l.Wrapping = fyne.TextWrapWord
			l2 := widget.NewLabel("Kliknij Dodaj aby wybrać wszystkie pliki w katalogu")
			l3 := widget.NewLabel("Kliknij w plik aby otworzyć podgląd")
			vBox.Add(l)
			vBox.Add(l2)
			vBox.Add(l3)
		} else {
			if loaddedFileForPreview == nil {
				return
			}
			o := toolbarLabel.ToolbarObject()
			var container *fyne.Container = o.(*fyne.Container)
			headerLabel := container.Objects[1].(*widget.Label)
			headerLabel.SetText("Podgląd: " + selectedPath)
			// headerLabel.Wrapping = fyne.TextWrapWord
			// headerLabel.Truncation = fyne.TextTruncateClip
			for _, element := range loaddedFileForPreview.Element {
				AddElement(vBox, &element)
			}
		}
	}
	return refreshCorpusPreviewFunc
}

func getCenterPanel() *fyne.Container {
	vBox := container.NewVBox()
	centerViewWithCorpusPreview = vBox // setting global reference, not ideal...
	toolbarLabel := NewToolbarLabel("Wybierz plik aby podejrzeć")
	topBar := widget.NewToolbar(
		toolbarLabel,
		widget.NewToolbarSpacer(),
		NewToolbarButton("Wybierz", func() {
			contains := slices.ContainsFunc(loadedFiles, func(e struct {
				path   string
				isFile bool
			}) bool {
				return e.path == selectedPath
			})
			if selectedPath != "" && !contains {
				stat, err := os.Stat(selectedPath)
				isFile := true
				if err == nil && stat.IsDir() {
					isFile = false
				}
				loadedFiles = append(loadedFiles, struct {
					path   string
					isFile bool
				}{selectedPath, isFile})
			}
			refreshListOfLoadedFiles()
		}),
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

	centerContainer := getCenterPanel()
	logo := canvas.NewImageFromResource(resourceCorpusreplacerlogoPng)
	logo.FillMode = canvas.ImageFillStretch
	logo.SetMinSize(fyne.NewSize(30, 30))

	topBar := widget.NewToolbar(
		widget.NewToolbarSpacer(),
		NewToolbarButton("Podumuj i zamień makra", func() {}),
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
