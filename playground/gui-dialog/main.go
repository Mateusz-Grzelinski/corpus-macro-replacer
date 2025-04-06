package corpus

import (
	"bytes"
	"fmt"
	"image/color"
	"io"
	"os"
	"os/exec"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func main() {
	main2()
}
func main2() {
	// Create a new Fyne application
	a := app.New()

	// Create a new window with a title
	w := a.NewWindow("Pop up demo by DaHerb")

	// Create a label widget with a message
	hello := widget.NewLabel("Demo Fyne show popup vs show modal popup")
	showPopUpButton := widget.NewButton("Show Pop-up", nil)

	// Define content and canvas objects for the pop-up
	content := canvas.NewText("This is the content of the pop-up",
		color.Black)

	// Declare a variable to hold the pop-up widget
	var popup *widget.PopUp

	// Define an action when the "Show Pop-up" button is tapped
	showPopUpButton.OnTapped = func() {
		// Create a container for the pop-up content, including a "Close" button
		popUpContent := container.NewVBox(
			content,
			widget.NewButton("Close", func() {
				popup.Hide() // Function to hide the pop-up
			}),
		)

		popup = widget.NewModalPopUp(popUpContent, w.Canvas())
		popup.Show()
	}

	// Set the main window's content to include the label and the Show-Pop-up button
	w.SetContent(container.NewVBox(
		hello,
		showPopUpButton,
	))
	// Show the main window and run the application
	w.ShowAndRun()
}

func main1() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Path and Output Dialog")
	myWindow.Resize(fyne.NewSize(600, 400))

	dialog.NewConfirm("asd", "asdadasdasd", func(ok bool) {

	}, myWindow)
	// Entry for path input
	pathEntry := widget.NewEntry()
	pathEntry.SetPlaceHolder("Enter a directory path")

	// Multi-line text widget for output
	output := widget.NewMultiLineEntry()
	output.Wrapping = fyne.TextWrapWord
	output.Disable()

	// Button to list files in the directory
	listFilesButton := widget.NewButton("List Files", func() {
		path := pathEntry.Text
		files, err := os.ReadDir(path)
		if err != nil {
			output.SetText(fmt.Sprintf("Error reading directory: %v", err))
			return
		}

		var fileList bytes.Buffer
		for _, file := range files {
			fileList.WriteString(file.Name() + "\n")
		}
		output.SetText(fileList.String())
	})

	// Confirmation button to run a command and capture stdout
	confirmButton := widget.NewButton("Run Command", func() {
		cmd := exec.Command("echo", "Hello from stdout!") // Example command
		var stdout bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = io.Discard

		if err := cmd.Run(); err != nil {
			output.SetText(fmt.Sprintf("Error executing command: %v", err))
			return
		}
		output.SetText(stdout.String())
	})

	// Layout with all elements
	content := container.NewVBox(
		widget.NewLabel("Enter a directory path and interact below:"),
		pathEntry,
		listFilesButton,
		confirmButton,
		widget.NewLabel("Output:"),
		output,
	)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}
