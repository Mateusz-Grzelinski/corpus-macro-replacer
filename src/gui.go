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
var SelectedPath string
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
		if !stat.IsDir() {
			elementFile, err := ReadCorpusFile(SelectedPath)
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

var macroIcon *canvas.Image = canvas.NewImageFromResource(resourceMacroSvg)

type MacroCacheKey struct {
	Path         string
	ElementIndex int
	ADIndex      int
	MacroIndex   int
}

type MacroContainer struct {
	widget.BaseWidget
	all      *fyne.Container
	header   *fyne.Container
	content  *fyne.Container
	isOpen   bool
	oldMakro *M1
	newMakro *M1
}

// NewMacroContainer creates a new instance of MacroContainer
func NewMacroContainer(objects ...fyne.CanvasObject) *MacroContainer {
	c := container.NewVBox(objects...)
	mc := &MacroContainer{
		all:     nil,
		header:  nil,
		content: c,
		isOpen:  false,
	}

	previewButton := widget.NewButtonWithIcon("", theme.VisibilityIcon(), func() {})
	previewButton.OnTapped = func() {
		if mc.isOpen {
			previewButton.Icon = theme.VisibilityIcon()
			c.Hide()
		} else {
			previewButton.Icon = theme.VisibilityOffIcon()
			c.Show()
		}
		previewButton.Refresh()
		mc.isOpen = !mc.isOpen
	}
	loadThisMacroButton := widget.NewButtonWithIcon("Zaczytaj to makro", theme.NavigateNextIcon(), func() {
		if mc.oldMakro == nil {
			return
		}
		macroGuessedPath := filepath.Join(MacrosDefaultPathNormal, mc.oldMakro.MakroName) + ".CMK"
		addToLoadedFilesAndRefresh(SelectedPath)
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
		c.Refresh()
		refreshCorpusPreviewFunc() // todo
	},
	)
	mc.header = container.NewBorder(nil, nil,
		previewButton,
		loadThisMacroButton,
	)
	mc.all = container.NewVBox(mc.header, mc.content)
	mc.content.Hide()

	mc.ExtendBaseWidget(mc)
	return mc
}

// func smartTextComparison(oldDAT string, newDAT string) (string, string) {
// 	var oldReformatted strings.Builder
// 	var newReformatted strings.Builder

// 	// load everything related to old
// 	oldVariablesKeys, oldValues, oldVariablesComments := loadValuesFromSection(oldDAT)

// 	// load everything related to new
// 	newVariablesKeys, newValues, newVariablesComments := loadValuesFromSection(newDAT)
// 	return oldReformatted.String(), newReformatted.String()
// }

func (mc *MacroContainer) SetNewMacro(newMakro *M1) {
	con := mc
	con.newMakro = newMakro
	con.header.Objects[0].(*widget.Button).SetText(newMakro.MakroName + " (do zamiany)")
	con.header.Refresh()
	{
		varijableText := con.content.Objects[0]
		UpdateMakro(con.oldMakro, con.newMakro, false)

		text := "`[VARIJABLE]`\n\n```\n" + strings.Join(decodeAllCMKLines(con.newMakro.Varijable.DAT), "\n") + "\n```"
		multilineNew := widget.NewRichTextFromMarkdown(text)
		multilineNew.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
		split := container.NewHSplit(varijableText, multilineNew)
		con.content.Objects[0] = split
		split.Refresh()
	}
	// jointText := con.header.Objects[1].(*widget.RichText)
}
func (mc *MacroContainer) SetOldMacro(oldMakro *M1) {
	con := mc
	con.oldMakro = oldMakro
	con.newMakro = nil
	con.header.Objects[0].(*widget.Button).SetText(oldMakro.MakroName)
	con.header.Refresh()
	con.content.RemoveAll() // not ideal...
	{
		textOld := "`[VARIJABLE]`\n\n```\n" + strings.Join(decodeAllCMKLines(oldMakro.Varijable.DAT), "\n") + "\n```"
		multilineOld := widget.NewRichTextFromMarkdown(textOld)
		multilineOld.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
		con.content.Add(multilineOld)
	}
	if oldMakro.Joint != nil {
		textOld := "`[JOINT]`\n\n```\n" + strings.Join(decodeAllCMKLines(cmp.Or(oldMakro.Joint.DAT, "")), "\n") + "\n```"
		multilineOld := widget.NewRichTextFromMarkdown(textOld)
		multilineOld.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
		con.content.Add(multilineOld)
	}
	var showAllMakro *widget.Button
	showAllMakro = widget.NewButton("Pokaż całe makro", func() {
		con.content.Remove(showAllMakro)
		if oldMakro.Formule != nil {
			textOld := "`[Formule]`\n\n```\n" + strings.Join(decodeAllCMKLines(cmp.Or(oldMakro.Formule.DAT, "")), "\n") + "\n```"
			multilineOld := widget.NewRichTextFromMarkdown(textOld)
			multilineOld.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
			con.content.Add(multilineOld)
		}
		for i, item := range oldMakro.Pocket {
			textOld := "`[POCKET" + strconv.Itoa(i) + "]`\n\n```\n" + strings.Join(decodeAllCMKLines(cmp.Or(item.DAT, "")), "\n") + "\n```"
			multilineOld := widget.NewRichTextFromMarkdown(textOld)
			multilineOld.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
			con.content.Add(multilineOld)
		}
		for i, item := range oldMakro.Potrosni {
			textOld := "`[POTROSNI" + strconv.Itoa(i) + "]`\n\n```\n" + strings.Join(decodeAllCMKLines(cmp.Or(item.DAT, "")), "\n") + "\n```"
			multilineOld := widget.NewRichTextFromMarkdown(textOld)
			multilineOld.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
			con.content.Add(multilineOld)
		}
		for i, item := range oldMakro.Grupa {
			textOld := "`[GRUPA" + strconv.Itoa(i) + "]`\n\n```\n" + strings.Join(decodeAllCMKLines(cmp.Or(item.DAT, "")), "\n") + "\n```"
			multilineOld := widget.NewRichTextFromMarkdown(textOld)
			multilineOld.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
			con.content.Add(multilineOld)
		}
		for i, item := range oldMakro.Raster {
			textOld := "`[RASTER" + strconv.Itoa(i) + "]`\n\n```\n" + strings.Join(decodeAllCMKLines(cmp.Or(item.DAT, "")), "\n") + "\n```"
			multilineOld := widget.NewRichTextFromMarkdown(textOld)
			multilineOld.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
			con.content.Add(multilineOld)
		}
		for i, item := range oldMakro.Makro {
			textOld := "`[MAKRO" + strconv.Itoa(i) + "]`\n\n```\n" + strings.Join(decodeAllCMKLines(cmp.Or(item.DAT, "")), "\n") + "\n```"
			multilineOld := widget.NewRichTextFromMarkdown(textOld)
			multilineOld.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
			con.content.Add(multilineOld)
		}
		con.content.Add(widget.NewButtonWithIcon("Zwiń", theme.VisibilityOffIcon(), func() {
			con.content.Hide()
			con.header.Objects[0].(*widget.Button).SetIcon(theme.VisibilityIcon())
			con.header.Refresh()
		}))
	})
	con.content.Add(showAllMakro)
	mc.Refresh()
}

func (mc *MacroContainer) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(mc.all)
}

// wanted to use cache so that accordion does not blink when refreshing, but does not work...
func getCachedMacroContainer(
	Path string,
	ElementIndex int,
	ADIndex int,
	MacroIndex int,
) *MacroContainer {
	key := MacroCacheKey{Path, ElementIndex, ADIndex, MacroIndex}
	con, found := macroCache[key]
	if found {
		return con
	}
	_con := NewMacroContainer()
	macroCache[key] = _con
	return _con
}

var macroCache map[MacroCacheKey](*MacroContainer) = map[MacroCacheKey](*MacroContainer){}

func RefreshMacro(con *MacroContainer, oldMakro *M1) {
	con.SetOldMacro(oldMakro)

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
	}
	if newMakro != nil {
		con.SetNewMacro(newMakro)
	}
}

// formatka/Daske
func RefreshPlate(con *fyne.Container, element *Element, elementIndex int, adIndex int) {
	daskeName := element.Daske.AD[adIndex].DName.Value
	numMakros := 0
	for _, spoj := range element.Elinks.Spoj {
		_adIndex, _ := strconv.Atoi(spoj.O1.Value)
		if _adIndex == adIndex {
			numMakros++
		}
	}
	accordionContent := container.NewVBox()

	for spojIndex, spoj := range element.Elinks.Spoj {
		_adIndex, _ := strconv.Atoi(spoj.O1.Value)
		if _adIndex != adIndex {
			continue
		}
		_con := getCachedMacroContainer(SelectedPath, elementIndex, spojIndex, adIndex)
		RefreshMacro(_con, &spoj.Makro1)
		accordionContent.Add(_con)
	}
	item1 := widget.NewAccordionItem("▧ Formatka: '"+daskeName+"' (makra: "+strconv.Itoa(numMakros)+")", accordionContent)
	con.Add(widget.NewAccordion(item1))
}

var cabinetIcon = canvas.NewImageFromResource(resourceCabinetSvg)

// todo make refresh instead add
func RefreshElement(con *fyne.Container, element *Element, elementIndex int) {
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
		_con := container.NewVBox()
		RefreshPlate(_con, element, elementIndex, adIndex)
		con.Add(_con)
	}
	con.Add(widget.NewSeparator())
}

func RefreshElementFile(con *fyne.Container) {
	for elementIndex, element := range loaddedFileForPreview.Element {
		// _con := container.NewVBox()
		RefreshElement(con, &element, elementIndex)
		// con.Add(_con)
	}
}

var refreshCorpusPreviewFunc func()

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
			if loaddedFileForPreview == nil {
				return
			}
			o := toolbarLabel.ToolbarObject()
			var container *fyne.Container = o.(*fyne.Container)
			headerLabel := container.Objects[1].(*widget.Label)
			headerLabel.SetText("Podgląd: " + SelectedPath)
			// headerLabel.Wrapping = fyne.TextWrapWord
			// headerLabel.Truncation = fyne.TextTruncateClip
			RefreshElementFile(vBox)
			// vBox.Add(getCachedMacroContainer("", 0, 0, 0))
			vBox.Refresh()
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

func getCenterPanel() *fyne.Container {
	vBox := container.NewVBox()
	centerViewWithCorpusPreview = vBox // setting global reference, not ideal...
	toolbarLabel := NewToolbarLabel("Wybierz plik aby podejrzeć")
	topBar := widget.NewToolbar(
		toolbarLabel,
		widget.NewToolbarSpacer(),
		NewToolbarButton("Zaczytaj ten plik/katalog", func() {
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

	centerContainer := getCenterPanel()
	logo := canvas.NewImageFromResource(resourceCorpusreplacerlogoPng)
	logo.FillMode = canvas.ImageFillStretch
	logo.SetMinSize(fyne.NewSize(30, 30))

	topBar := widget.NewToolbar(
		widget.NewToolbarSpacer(),
		NewToolbarButton("Podsumuj i zamień makra", func() {}),
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
