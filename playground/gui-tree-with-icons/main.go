package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"

	xWidget "fyne.io/x/fyne/widget"
)

type SelectableFileTree struct {
	xWidget.FileTree
	Selected map[widget.TreeNodeID]bool
	// Checks map[widget.TreeNodeID]*widget.Check
}

func NewSelectableFileTree(root fyne.URI) *SelectableFileTree {
	tree := &SelectableFileTree{
		FileTree: *xWidget.NewFileTree(root),
		Selected: map[widget.TreeNodeID]bool{},
	}
	// oldCreateNode := tree.CreateNode
	// tree.CreateNode = func(branch bool) fyne.CanvasObject {
	// 	border := oldCreateNode(branch)
	// 	check := widget.NewCheck("", nil)
	// 	// tree.Checks[] = check
	// 	// todo if branch then add 3 state checkbox
	// 	return container.NewHBox(check, border)
	// }
	// oldUpdateNode := tree.UpdateNode
	// tree.UpdateNode = func(id widget.TreeNodeID, branch bool, node fyne.CanvasObject) {
	// 	hbox := node.(*fyne.Container)
	// 	check := hbox.Objects[0].(*widget.Check)
	// 	check.OnChanged = func(b bool) {
	// 		tree.Selected[id] = b
	// 	}
	// 	check.Checked = tree.Selected[id]
	// 	check.Refresh()
	// 	// tree.Refresh()
	// 	oldUpdateNode(id, branch, hbox.Objects[1])
	// }
	// tree.OnSelected = func(uid widget.TreeNodeID) {
	// 	tree.Selected[uid] = true
	// 	// tree.RefreshItem()
	// }
	tree.ExtendBaseWidget(tree)

	return tree
}

// func (t *SelectableFileTree) Select(uid widget.TreeNodeID) {
// 	fmt.Printf(uid)
// 	// t.Selected = append(t.Selected, uid)
// 	// t.FileTree.Select(uid)
// }

// func (t *SelectableFileTree) RefreshItem(id widget.TreeNodeID) {
// }

func main() {
	main2()
	// a := app.NewWithID("demo.selectableFileTree")
	// window := a.NewWindow("Disk Tree")
	// window.Resize(fyne.NewSize(200, 400))

	// lightIcon := fyne.NewStaticResource("light_icon", theme.FolderNewIcon().Content())

	// customTheme := NewCustomTheme(theme.DefaultTheme())
	// customTheme.SetIcon("custom-icon", lightIcon)
	// a.Settings().SetTheme(customTheme)
	// icon := widget.NewIcon(customTheme.Icon("custom-icon"))
	// selected := []string{}

	// var fileTree *SelectableFileTree = NewSelectableFileTree(storage.NewFileURI(`C:\`))
	// fmt.Print(fileTree)
	// fileTree = xWidget.NewFileTree(storage.NewFileURI(`C:\`))
	// fileTree.OnSelected = func(uid widget.TreeNodeID) {
	// 	selected = append(selected, uid)
	// 	fileTree.Refresh()
	// }
	// // var check1 *widget.Check = widget.NewCheck("original", nil)
	// var check2 *CheckN
	// check2 = NewCheckN("adasd", func(b bool) {})
	// check2.Enable()
	// // func(b bool) {
	// // 	check2.Checked = b
	// // 	// check2.Refresh()
	// // })
	// check2.SetChecked(false)
	// check2.SetText("ola!")

	// var border *fyne.Container = container.NewBorder(nil, nil, nil,
	// 	nil,
	// 	container.NewVBox(check2),
	// )

	// window.SetContent(border)
	// window.SetContent(container.NewVBox(check2))
	// window.Resize(fyne.NewSize(1000, 700))
	// window.ShowAndRun()
}
