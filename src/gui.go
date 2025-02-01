package main

import (
	"cmp"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/cmd/fyne_settings/settings"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	xWidget "fyne.io/x/fyne/widget"
)

const (
	CorpusMacroReplacerDefaultPath = `C:\Tri D Corpus\CorpusMacroReplacer\`
	CorpusFileBrowserDefaultPath   = `file://C:\Tri D Corpus\Corpus 5.0\elmsav\`
	MacrosDefaultPath              = `file://C:\Tri D Corpus\Corpus 5.0\Makro\`
)

var corpusMimeType = []string{"text/plain"}
var corpustExtensions = []string{".s3d", ".e3d"} // use with strings.Fold

var CorpusTypicallyUsedFolders = []string{
	`C:\Tri D Corpus\Corpus 4.0\elmsav\`,
	`C:\Tri D Corpus\Corpus 4.0\sobasav\`,
	`C:\Tri D Corpus\Corpus 5.0\elmsav\`,
	`C:\Tri D Corpus\Corpus 5.0\sobasav\`,
	`C:\Tri D Corpus\Corpus 6.0\elmsav\`,
	`C:\Tri D Corpus\Corpus 6.0\sobasav\`,
}

var loadedS3DFileForPreview *ProjectFile
var loadedE3DFileForPreview *ElementFile
var loadedFileForPreviewError error
var SelectedPath string
var MakroCollectionCache MakroCollection = MakroCollection{}
var corpusPreviewContainer *ElementFileContainer
var ListOfLoadedFilesContainer *fyne.Container
var MacrosToChangeEntries []*widget.Entry
var MacrosToChangeNamesEntries []*widget.Entry
var AddMakroButton *widget.Button
var DialogSizeDefault fyne.Size = fyne.NewSize(950, 650)

var refreshCorpusPreviewFunc func()

var loadedFiles []struct {
	path   string
	isFile bool
} = []struct {
	path   string
	isFile bool
}{}

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
			projectFile, elementFile, err := NewCorpusFile(SelectedPath)
			if elementFile != nil {
				loadedS3DFileForPreview = nil
				loadedE3DFileForPreview = elementFile
			} else if projectFile != nil {
				loadedS3DFileForPreview = projectFile
				loadedE3DFileForPreview = nil
			}
			if err != nil {
				log.Println(err)
			}
			loadedFileForPreviewError = err
		}
	}
	if refreshCorpusPreviewFunc != nil {
		refreshCorpusPreviewFunc()
	}
}

func getLeftPanel(a fyne.App, myWindow *fyne.Window) *fyne.Container {
	var CorpusFileTreeContainer *fyne.Container
	var fileTree *xWidget.FileTree

	nothingOpenLabel := widget.NewLabel("Brak otwartych plików!")
	nothingOpenLabel.Truncation = fyne.TextTruncateClip
	nothingOpenLabel.Alignment = fyne.TextAlignCenter

	typicalLabel := widget.NewLabel("Typowe ścieżki")
	typicalLabel.Truncation = fyne.TextTruncateClip
	typicalLabel.Alignment = fyne.TextAlignCenter

	defaultContainer := container.NewVBox(nothingOpenLabel, typicalLabel)
	for _, p := range CorpusTypicallyUsedFolders {
		_, err := os.Stat(p)
		if err == nil {
			defaultContainer.Add(widget.NewButtonWithIcon(p, theme.FolderOpenIcon(), func() {
				// pretty fat finger solution, duplicated below...
				CorpusFileTreeContainer.Remove(fileTree)
				fileTree = xWidget.NewFileTree(storage.NewFileURI(p))
				fileTree.OnSelected = CorpusFileTreeOnSelected
				// todo filter extension but allow directory
				// fileTree.Filter = storage.NewExtensionFileFilter([]string{".S3D", ".E3D"})
				defaultContainer.RemoveAll()
				CorpusFileTreeContainer.Remove(defaultContainer)
				CorpusFileTreeContainer.Add(fileTree)
				CorpusFileTreeContainer.Refresh()
			}))
		}
	}
	recentlyUsedLabel := widget.NewLabel("Ostatnio otwarte")
	recentlyUsedLabel.Truncation = fyne.TextTruncateClip
	recentlyUsedLabel.Alignment = fyne.TextAlignCenter

	defaultContainer.Add(recentlyUsedLabel)
	recentlyUsedFolders := a.Preferences().StringList("recentlyUsedFodlers")
	slices.Reverse(recentlyUsedFolders)
	for _, p := range recentlyUsedFolders {
		l := widget.NewLabel(p)
		l.Wrapping = fyne.TextWrapBreak
		defaultContainer.Add(container.NewBorder(nil, nil,
			// todo duplicated
			widget.NewButtonWithIcon("", theme.FolderOpenIcon(), func() {
				// pretty fat finger solution, duplicated below...
				CorpusFileTreeContainer.Remove(fileTree)
				fileTree = xWidget.NewFileTree(storage.NewFileURI(p))
				fileTree.OnSelected = CorpusFileTreeOnSelected
				fileTree.Filter = storage.NewMimeTypeFileFilter(corpusMimeType)
				defaultContainer.RemoveAll()
				CorpusFileTreeContainer.Remove(defaultContainer)
				CorpusFileTreeContainer.Add(fileTree)
				CorpusFileTreeContainer.Refresh()
			}), nil, l),
		)
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
			fileTree.Filter = storage.NewExtensionFileFilter(corpustExtensions)
			defaultContainer.RemoveAll()
			CorpusFileTreeContainer.Remove(defaultContainer)
			CorpusFileTreeContainer.Add(fileTree)
			CorpusFileTreeContainer.Refresh()

			lastList := a.Preferences().StringList("recentlyUsedFodlers")
			lastList = append(lastList, lu.Path())
			if len(lastList) > 10 {
				lastList = lastList[1:]
			}
			a.Preferences().SetStringList("recentlyUsedFodlers", lastList)
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

	wczytajPlikKatalogButton := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		addToLoadedFilesAndRefresh(SelectedPath)
	})
	CorpusFileTreeContainer = container.NewBorder(
		container.NewBorder(nil, nil, nil, wczytajPlikKatalogButton, openFilesButton),
		nil, nil, nil, container.NewVScroll(defaultContainer))

	files := widget.NewAccordionItem("Pliki Corpusa", CorpusFileTreeContainer)
	files.Open = true
	return CorpusFileTreeContainer
}

// best effort, returned path might not exist
func getMacroNameByFileName(makroSearchPath string, path string, cache *MakroCollection) string {
	if cache != nil {
		if out := cache.GetMacroNameByFileName(path); out != nil {
			return *out
		}
	}
	relPath, err := filepath.Rel(makroSearchPath, path)
	if err != nil {
		path = relPath
	}
	base, found := strings.CutSuffix(path, ".CMK")
	if found {
		return filepath.Base(base)
	}
	return filepath.Base(path)
}

func RemoveEntry(slice []*widget.Entry, item *widget.Entry) []*widget.Entry {
	for i, o := range slice {
		if o != item {
			continue
		}

		removed := make([]*widget.Entry, len(slice)-1)
		copy(removed, slice[:i])
		copy(removed[i:], slice[i+1:])

		return removed
	}
	return slice
}

func getRightPanel(a fyne.App, myWindow *fyne.Window) *widget.Accordion {
	macrosToChangeContainer := container.NewVBox()

	AddMakroButton = widget.NewButton("Dodaj makro do zamiany", func() {
		newMacroNameEntry := widget.NewEntry()
		newMacroNameEntry.PlaceHolder = "Nazwa makra: Nawierty_uniwersalne_28mm"
		autoFixError := widget.NewButton("AutoFix", nil)
		autoFixError.Hide()
		makroErrorLabel := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		makroErrorLabel.Hide()
		makroErrorLabel.Wrapping = fyne.TextWrapBreak
		newMacroPathEntry := widget.NewEntry()
		newMacroPathEntry.SetPlaceHolder(`C:\Tri D Corpus\Corpus 5.0\Makro\*.CMK`)
		newMacroPathEntry.OnChanged = func(path string) {
			// wasteful but the best error reporting
			makroRootPath := fyne.CurrentApp().Preferences().String("makroSearchPath")
			_, err := NewMakroFromCMKFile(nil, path, &makroRootPath, MakroCollectionCache.GetMakroMappings())
			if err != nil {
				log.Printf("ERROR: reading makro failed: %s\n", err)
				makroErrorLabel.SetText(fmt.Sprintf("ERROR: %s", err))
				var targetErr *fs.PathError
				if errors.As(err, &targetErr) {
					if targetErr.Op == "open" && errors.Is(targetErr.Err, syscall.ERROR_FILE_NOT_FOUND) {
						autoFixError.Show()
						autoFixError.OnTapped = func() {
							card1Result := widget.NewLabel("")
							card1Result.Wrapping = fyne.TextWrapBreak
							card1Result.Importance = widget.DangerImportance
							makroSearchPath := a.Preferences().String("makroSearchPath")
							card1 := widget.NewCard("1. Szukam brakującego pliku w Makrach", "Szukam w "+makroSearchPath, container.NewVBox(card1Result))
							card2Result := widget.NewLabel("")
							card2Result.Wrapping = fyne.TextWrapBreak
							card2Result.Importance = widget.DangerImportance
							card2 := widget.NewCard("2. Odtwarzam brakujące makro z pliku Corpusa", "Szukam w "+cmp.Or(SelectedPath, ""), container.NewVBox(card2Result))
							card2.Hide()
							cardFail := widget.NewCard("Nie udało się naprawić", "", container.NewVBox(
								widget.NewRichTextFromMarkdown(`
- zaznacz inny plik Corpusa uruchom AutoFix jeszcze raz
- autor mógł zmienić nazwę pliku CMK (duże i małe litery mają znaczenie)
- autor mógł zmienić nazwę makra w pliku MakroCollection.dat (do edycji z poziomy corpusa)
- plik może już nie istnieć
- edycja nazwy makra nie jest jeszcze wspierana`),
							))
							cardFail.Hide()
							cardOk := widget.NewCard("Udało się naprawić", "", container.NewVBox(
								widget.NewRichTextFromMarkdown("Makro może mieć niepożadaną zawartość, polezam na nie spojrzeć"),
								widget.NewRichTextFromMarkdown("Jeszcze inne błędy mogą wystąpić, klikaj AutoFix do skutku"),
							))
							cardOk.Hide()
							filterDialog := dialog.NewCustom("Szukam brakującego makra", "Zamknij", container.NewVBox(card1, card2, cardFail, cardOk), *myWindow)
							continueRecovery := true
							baseFilename := filepath.Base(targetErr.Path)
							if continueRecovery {
								foundFile, _ := findFile(makroSearchPath, baseFilename)
								if foundFile != "" {
									err := copyFile(foundFile, targetErr.Path)
									if err != nil {
										card1Result.SetText(fmt.Sprintf("Znaleziono: \"%s\" i ale nie udało się skopiować do \"%s\": %s", foundFile, targetErr.Path, err))
									} else {
										card1Result.SetText(fmt.Sprintf("Znaleziono: \"%s\" i skopiowano do \"%s\"", foundFile, targetErr.Path))
										card1Result.Importance = widget.HighImportance
									}
									continueRecovery = false
								} else {
									card1Result.SetText(fmt.Sprintf("Nie znaleziono: \"%s\" w \"%s\"", baseFilename, makroSearchPath))
								}
							}

							if continueRecovery {
								var makros *ElementFile
								if loadedE3DFileForPreview != nil {
									makros = loadedE3DFileForPreview
								} else if loadedS3DFileForPreview != nil {
									makros = &loadedS3DFileForPreview.ElementFile
								}
								if makros != nil {
									makroSearchPath := a.Preferences().String("makroSearchPath")
									makroNameToFind := getMacroNameByFileName(makroSearchPath, targetErr.Path, &MakroCollectionCache)
								outerLoop:
									for _, e := range makros.Element {
										for _, s := range e.Elinks.Spoj {
											if s.Makro1.MakroName != makroNameToFind {
												continue
											}
											f, err := os.Create(targetErr.Path)
											if err != nil {
												card2Result.SetText(fmt.Sprintf("otwarty plik Corpusa ma makro \"%s\", ale nie można go zapisać: %s", baseFilename, err))
											} else {
												_, err1 := f.Write([]byte(fmt.Sprintf("// odzyskano z %s\n", SelectedPath)))
												err := s.Makro1.Save(f)
												if err1 != nil && err != nil {
													card2Result.SetText(fmt.Sprintf("odzyskano Makro \"%s\" z Corpusa, ale wystąpił błąd przy zapisywaniu: %s", baseFilename, err))
												} else {
													card2Result.SetText(fmt.Sprintf("odzyskano Makro \"%s\" z Corpusa i zapisano: %s. Zawartość pliku może być stara.", baseFilename, targetErr.Path))
													card2Result.Importance = widget.HighImportance
													continueRecovery = false
												}
											}
											break outerLoop
										}
									}
									if continueRecovery {
										card2Result.SetText(fmt.Sprintf("nie znaleziono Makra \"%s\" w otwartym pliku Corpusa", baseFilename))
									}
								} else {
									card2Result.SetText(fmt.Sprintf("otwórz plik corpusa aby poszukać w nim Makra: \"%s\"", baseFilename))
								}
								card2.Show()
							}
							if continueRecovery {
								cardFail.Show()
							} else {
								makroErrorLabel.Hide()
								autoFixError.Hide()
								cardOk.Show()
							}

							filterDialog.Resize(DialogSizeDefault)
							filterDialog.Show()
						}
					}
				}
				makroErrorLabel.Importance = widget.DangerImportance
				makroErrorLabel.Show()
			} else {
				makroSearchPath := a.Preferences().String("makroSearchPath")
				newMacroNameEntry.SetText(getMacroNameByFileName(makroSearchPath, path, &MakroCollectionCache))
				makroErrorLabel.Importance = widget.MediumImportance
				makroErrorLabel.Hide()
				refreshCorpusPreviewFunc()
			}
			makroErrorLabel.Refresh()
		}
		MacrosToChangeEntries = append(MacrosToChangeEntries, newMacroPathEntry)           // save to global var
		MacrosToChangeNamesEntries = append(MacrosToChangeNamesEntries, newMacroNameEntry) // save global var
		var row *fyne.Container
		row = container.NewBorder(newMacroNameEntry, container.NewVBox(container.NewBorder(nil, nil, nil, autoFixError, makroErrorLabel), widget.NewSeparator()), nil,
			container.NewHBox(
				widget.NewButtonWithIcon("", theme.FolderOpenIcon(), func() {
					fileOpenDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
						if err != nil {
							dialog.ShowError(err, *myWindow)
							return
						}
						if reader == nil {
							return
						}
						path := reader.URI().Path()
						newMacroPathEntry.SetText(path)

						makroSearchPath := a.Preferences().String("makroSearchPath")
						newMacroNameEntry.SetText(getMacroNameByFileName(makroSearchPath, path, &MakroCollectionCache))
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
					MacrosToChangeNamesEntries = RemoveEntry(MacrosToChangeNamesEntries, newMacroNameEntry)
					MacrosToChangeEntries = RemoveEntry(MacrosToChangeEntries, newMacroPathEntry)
					macrosToChangeContainer.Refresh()
					refreshCorpusPreviewFunc()
				}),
			),
			newMacroPathEntry,
		)
		macrosToChangeContainer.Add(row)
	})
	AddMakroButton.OnTapped()

	item1 := widget.NewAccordionItem("Makra do zamiany",
		container.NewScroll(
			container.NewVBox(
				AddMakroButton,
				macrosToChangeContainer,
			),
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

func isCorpusExtension(path string) bool {
	// todo use strings.EqualFold()?
	ext := strings.ToLower(filepath.Ext(path))
	return slices.Contains(corpustExtensions, ext)
}

func generateRefreshCorpusPreview(a fyne.App, vBox *fyne.Container, toolbarLabel *toolbarLabel) func() {
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
			l2 := widget.NewLabel(`Kliknij "+Wczytaj plik/katalog" aby wybrać WSZYTKIE pliki w katalogu`)
			l3 := widget.NewLabel("Kliknij plik aby otworzyć podgląd")
			vBox.Add(l)
			vBox.Add(l2)
			vBox.Add(l3)
		} else {
			if !isCorpusExtension(SelectedPath) {
				l := widget.NewLabel("To nie jest plik Corpusa: ")
				l.Wrapping = fyne.TextWrapWord
				lErr := widget.NewLabel(SelectedPath)
				lErr.Wrapping = fyne.TextWrapWord
				vBox.Add(l)
				vBox.Add(lErr)
				return
			}
			if loadedFileForPreviewError != nil {
				l := widget.NewLabel("Błąd podczas czytania Corpusa:")
				l.Wrapping = fyne.TextWrapWord
				lErr := widget.NewLabel(fmt.Sprintf("%s", loadedFileForPreviewError))
				lErr.Wrapping = fyne.TextWrapWord
				vBox.Add(l)
				vBox.Add(lErr)
				return
			}
			if loadedE3DFileForPreview == nil && loadedS3DFileForPreview == nil {
				return
			}
			var elementFileToDisplay *ElementFile
			if loadedE3DFileForPreview != nil {
				elementFileToDisplay = loadedE3DFileForPreview
			}
			if loadedS3DFileForPreview != nil {
				elementFileToDisplay = &loadedS3DFileForPreview.ElementFile
			}

			o := toolbarLabel.ToolbarObject()
			var container *fyne.Container = o.(*fyne.Container)
			headerLabel := container.Objects[1].(*widget.Label)
			headerLabel.SetText("Podgląd: " + SelectedPath)
			compact := a.Preferences().Bool("compact")
			hideElementsWithZeroMacros := a.Preferences().Bool("hideElementsWithZeroMacros")
			if corpusPreviewContainer == nil || corpusPreviewContainer.elementFile != elementFileToDisplay {
				elemFileCon := NewElementFileContainer(elementFileToDisplay, compact, hideElementsWithZeroMacros)
				corpusPreviewContainer = elemFileCon
				vBox.Add(elemFileCon)
			} else {
				// corpusPreviewContainer.elementFile = elementFileToDisplay
				corpusPreviewContainer.Update(elementFileToDisplay, compact, hideElementsWithZeroMacros)
				vBox.Add(corpusPreviewContainer)
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

func getCenterPanel(a fyne.App, w fyne.Window) *fyne.Container {
	vBox := container.NewVBox(NewCorpusMakroReplacerSettings(a))
	// centerViewWithCorpusPreview = vBox // setting global reference, not ideal...
	toolbarLabel := NewToolbarLabel("Wybierz plik aby podejrzeć")
	topBar := widget.NewToolbar(
		toolbarLabel,
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.Icon("filter"), func() {
			checkHideZeroMacrosElements := widget.NewCheck("Pokazuj tylko elementy z przynajmniej jednym makrem", func(b bool) {
				a.Preferences().SetBool("hideElementsWithZeroMacros", b)
			})
			checkHideZeroMacrosElements.Checked = a.Preferences().BoolWithFallback("hideElementsWithZeroMacros", true)
			checkCompact := widget.NewCheck("Kompaktowy widok", func(b bool) {
				a.Preferences().SetBool("compact", b)
			})
			checkCompact.Checked = a.Preferences().Bool("compact")

			filterDialog := dialog.NewCustom("Filtruj", "Ok", container.NewVBox(
				checkHideZeroMacrosElements,
				checkCompact,
			), w)

			filterDialog.Show()
			filterDialog.SetOnClosed(func() {
				refreshCorpusPreviewFunc()
			})
		}),
		NewToolbarButtonWithIcon("Wczytaj plik/katalog", theme.ContentAddIcon(), func() {
			addToLoadedFilesAndRefresh(SelectedPath)
		},
		),
	)
	generateRefreshCorpusPreview(a, vBox, toolbarLabel)

	scrollable := container.NewScroll(vBox)

	return container.NewBorder(topBar, nil, nil, nil, scrollable)
}

func parseURL(urlStr string) *url.URL {
	link, err := url.Parse(urlStr)
	if err != nil {
		fyne.LogError("Could not parse URL", err)
	}

	return link
}

func RunGui() {
	a := app.NewWithID("pl.net.stolarz")
	myWindow := a.NewWindow("Corpus Makro Replacer")
	myWindow.SetIcon(resourceCorpusreplacerlogoPng)
	customTheme := NewCustomTheme(theme.DefaultTheme())
	a.Settings().SetTheme(customTheme)

	centerContainer := getCenterPanel(a, myWindow)
	logo := canvas.NewImageFromResource(resourceCorpusreplacerlogoPng)
	logo.FillMode = canvas.ImageFillStretch
	logo.SetMinSize(fyne.NewSize(30, 30))

	var outputButton widget.ToolbarItem

	openSettings := func() {
		w := a.NewWindow("Ustawienia")
		settingsContent := settings.NewSettings().LoadAppearanceScreen(w)
		makroSettings := NewCorpusMakroReplacerSettings(a)
		w.SetContent(
			container.NewVBox(makroSettings, settingsContent),
		)
		w.Resize(fyne.NewSize(440, 520))
		w.SetIcon(resourceCorpusreplacerlogoPng)
		w.Show()
	}
	showAbout := func() {
		w := a.NewWindow("O programie")
		w.SetIcon(resourceCorpusreplacerlogoPng)
		w.SetContent(container.NewVBox(
			widget.NewLabel(fmt.Sprintf("Version: %s", Version)),
			widget.NewLabel("Author: Mateusz Grzeliński"),
			widget.NewHyperlink("Source - documentation - license", parseURL("https://github.com/Mateusz-Grzelinski/corpus-macro-replacer.git")),
		))
		w.Show()
	}
	aboutItem := widget.NewToolbarAction(resourceCorpusreplacerlogoPng, showAbout)
	settingsAction := widget.NewToolbarAction(theme.Icon("gear"), openSettings)

	outputButton = NewToolbarButtonWithIcon("Podsumuj i zamień makra", theme.MediaPlayIcon(), onTappedOutputPopup(a, outputButton, myWindow))
	topBar := widget.NewToolbar(
		aboutItem,
		settingsAction,
		widget.NewToolbarSpacer(),
		outputButton,
	)
	left := getLeftPanel(a, &myWindow)
	right := getRightPanel(a, &myWindow)

	// collection, _ := LoadMakroCollection(inputPath)
	// MakroCollectionCache = &collection

	hSplit := container.NewHSplit(left, centerContainer)
	hSplit.SetOffset(0.2)
	center := container.NewHSplit(hSplit, right)
	center.SetOffset(0.8)
	var border *fyne.Container = container.NewBorder(container.NewBorder(nil, nil, nil, nil, topBar), nil, nil, nil, center)

	myWindow.SetContent(border)
	myWindow.Resize(fyne.NewSize(1000, 700))
	myWindow.ShowAndRun()
}
