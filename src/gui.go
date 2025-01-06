package main

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	xWidget "fyne.io/x/fyne/widget"
)

const CorpusFileBrowserDefaultPath = `file://C:\Tri D Corpus\Corpus 5.0\elmsav\`
const CorpusFileBrowserDefaultPath2 = `C:\Tri D Corpus\Corpus 5.0\elmsav\_modifications\ui\`
const MacrosFileBrowserDefaultPath = `C:\Tri D Corpus\Corpus 5.0\Makro\`

var loaddedFile *ElementFile

func CorpusFileTreeOnSelected(uid widget.TreeNodeID) {
	parsedURL, err := url.Parse(string(uid))
	if parsedURL.Scheme != "file" {
		log.Println("URL scheme must be 'file'")
	}
	path := parsedURL.Path
	// path := storage.NewFileURI(uid).Path()
	elementFile, err := ReadCorpusFile(path)
	loaddedFile = elementFile
	if err != nil {
		log.Println(err)
	}
}

func getLeftPanel(myWindow *fyne.Window) *fyne.Container {
	// nothingOpen := widget.NewLabel("No open files")
	nothingOpen := xWidget.NewFileTree(storage.NewFileURI(CorpusFileBrowserDefaultPath2))
	nothingOpen.OnSelected = CorpusFileTreeOnSelected
	var CorpusFileTreeContainer *fyne.Container
	var openFilesButton *widget.Button = widget.NewButton("Otwórz katalog", func() {
		folderDialog := dialog.NewFolderOpen(func(lu fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, *myWindow)
				return
			}
			fmt.Println(lu)
			if lu == nil {
				return
			}
			fileTree := xWidget.NewFileTree(storage.NewFileURI(lu.Path()))
			CorpusFileTreeContainer.Remove(nothingOpen)
			CorpusFileTreeContainer.Add(fileTree)
			CorpusFileTreeContainer.Refresh()

		}, *myWindow)

		uri, err := storage.ParseURI(CorpusFileBrowserDefaultPath)
		if err != nil {
			listable, err := storage.ListerForURI(uri)
			if err != nil {
				folderDialog.SetLocation(listable)
			}
		}
		folderDialog.Show()
	})
	CorpusFileTreeContainer = container.NewBorder(openFilesButton, nil, nil, nil, nothingOpen)

	files := widget.NewAccordionItem("Pliki Corpusa", CorpusFileTreeContainer)
	files.Open = true
	// ---------------------------------------------------------------------------------------

	// macroNothingOpen := xWidget.NewFileTree(storage.NewFileURI(MacrosFileBrowserDefaultPath))
	// macros := widget.NewAccordionItem("Pliki Makro", macroNothingOpen)
	// macros.Open = true

	// accordion := widget.NewAccordion(
	// files,
	// macros,
	// )
	// accordion.MultiOpen = true // Allow multiple items to be open at the same time
	// return container.NewStack(
	// 	accordion,
	// )
	return CorpusFileTreeContainer
}

func getRightPanel(myWindow *fyne.Window) *widget.Accordion {
	item1 := widget.NewAccordionItem("Marka do zamiany", container.NewVBox(
		container.NewBorder(nil, nil, nil,
			widget.NewIcon(theme.DeleteIcon()),
			// container.NewHBox(
			widget.NewEntry(),
			// ),
			// widget.NewLabel("->"),
			widget.NewEntry(),
		),
	// )
	))
	item1.Open = true
	// item2 := widget.NewAccordionItem("Edytowane makra", widget.NewLabel("Szafki"))
	// item3 := widget.NewAccordionItem("Filtruj formatki", widget.NewLabel("Formatki"))
	// item4 := widget.NewAccordionItem("Ustawienia", widget.NewLabel("Formatki"))
	accordion := widget.NewAccordion(item1)
	accordion.MultiOpen = true
	return accordion
}

func AddMacro(con *fyne.Container, makro *M1) {
	con.Add(container.NewHBox(
		widget.NewCheck("", func(checked bool) {
			return
		}),
		// widget.NewIcon(theme.MenuDropDownIcon()),
		widget.NewRichTextFromMarkdown(makro.MakroName),
		// widget.NewIcon(theme.VisibilityIcon()),
		widget.NewIcon(theme.DocumentSaveIcon()),
		widget.NewIcon(theme.DocumentCreateIcon()),
	))

	con.Add(widget.NewLabel("[VARIJABLE]"))
	multiline := widget.NewMultiLineEntry()
	multiline.SetText(strings.Join(decodeAllCMKLines(makro.Varijable.DAT), "\n"))
	con.Add(multiline)
	// container.NewHSplit(
	// container.NewVBox(
	// multiline,
	// ),
	// container.NewVBox(
	// 	widget.NewRichTextFromMarkdown(`name=2`),
	// ),
	// ),
	// }
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
	item1 := widget.NewAccordionItem("Formatka: '"+daskeName+"' (makra: "+strconv.Itoa(numMakros)+")", vBox)
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

func AddElement(con *fyne.Container, element *Element) {
	// spacer := widget.NewLabel("")
	// plate := getPlate()
	// titleCheckbox := widget.NewCheck("Enable", nil)

	// // Create a label for the accordion title
	// titleLabel := widget.NewLabel("Custom Title")

	// // Create two icons
	// icon1 := widget.NewIcon(theme.ContentAddIcon())
	// icon2 := widget.NewIcon(theme.ContentRemoveIcon())

	// Combine the checkbox, label, and icons in a horizontal box
	// customTitle := container.NewHBox(
	// 	titleCheckbox,
	// 	titleLabel,
	// 	widget.NewSeparator(), // Spacer to push icons to the right
	// 	icon1,
	// 	icon2,
	// )
	// accordionItem := widget.NewAccordionItem("", widget.NewLabel("Accordion Content"))
	// accordionItem.Title = customTitle
	con.Add(
		container.NewHBox(
			// widget.NewIcon(theme.FileIcon()),
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

func getCenterPanel() *fyne.Container {
	vBox := container.NewVBox()
	toolbarLabel := NewToolbarLabel("Load file...")
	topBar := widget.NewToolbar(
		toolbarLabel,
		widget.NewToolbarSpacer(),
		NewToolbarButton("refresh", func() {
			if loaddedFile == nil {
				return
			}
			vBox.RemoveAll()
			o := toolbarLabel.ToolbarObject()
			var container *fyne.Container = o.(*fyne.Container)
			headerLabel := container.Objects[1].(*widget.Label)
			headerLabel.SetText("Plik: " + loaddedFile.FILE.Value)
			fmt.Print(o)
			for _, element := range loaddedFile.Element {
				AddElement(vBox, &element)
			}
		}),
	)

	scrollable := container.NewScroll(
		vBox,
	)
	var centerPanel *fyne.Container = container.NewBorder(topBar, nil, nil, nil, scrollable)

	return centerPanel
}

func RunGui() {
	myApp := app.NewWithID("pl.net.stolarz")
	myWindow := myApp.NewWindow("Corpus Makro Replacer")
	// vBox := container.NewVBox(
	// 	container.NewHBox(
	// 		widget.NewIcon(theme.FileIcon()),
	// 		widget.NewRichTextFromMarkdown("# lewy_gorny.E3D"),
	// 	),
	// )
	// AddCabinet(vBox)
	// AddCabinet(vBox)
	// AddCabinet(vBox)
	// AddCabinet(vBox)
	// AddCabinet(vBox)
	// AddCabinet(vBox)

	// scrollable := container.NewScroll(
	// 	vBox,
	// )
	// fmt.Print(scrollable)
	// scrollable.SetMinSize(fyne.NewSize(500, 600))
	// centerContainer :=
	// 	container.NewVBox(
	// 		// container.NewHBox(
	// 		// 	widget.NewIcon(theme.FileIcon()),
	// 		widget.NewRichTextFromMarkdown("# File: Stolarz.E3D"),
	// 		// ),
	// 		// scrollable,

	// 	)
	centerContainer := getCenterPanel()

	topBar := widget.NewToolbar(
		widget.NewToolbarSpacer(),
		NewToolbarButton("Zamień Makra", func() {}),
	)
	fmt.Print(topBar)

	left := getLeftPanel(&myWindow)
	right := getRightPanel(&myWindow)
	hSplit := container.NewHSplit(left, centerContainer)
	hSplit.SetOffset(0.2)
	center := container.NewHSplit(hSplit, right)
	center.SetOffset(0.8)
	var border *fyne.Container = container.NewBorder(topBar, nil, nil, nil, center)

	myWindow.SetContent(border)
	myWindow.Resize(fyne.NewSize(1000, 700))
	myWindow.ShowAndRun()
}
