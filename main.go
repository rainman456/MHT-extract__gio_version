package main

import (
	"fmt"
	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
    "gioui.org/op/clip"
	"gioui.org/op/paint"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"github.com/ncruces/zenity"
	"mhtmlExtractor/mhtmlparser" 
)

type (
	C = layout.Context
	D = layout.Dimensions
)

// Resource represents an extracted resource from an MHTML file.
type Resource struct {
	Type     string
	Filename string
	Size     int64
	Source   string
	Selected bool
}

type MHTMLApp struct {
	window           *app.Window
	theme            *material.Theme
	darkMode         bool
	filePath         widget.Editor
	browseBtn        widget.Clickable
	selectAllBtn     widget.Clickable
	darkModeBtn      widget.Clickable
	extractBtn       widget.Clickable
	outputDirBtn     widget.Clickable
	fetchExternalBtn widget.Bool
	rawContent       widget.Editor
	status           string
	selectedFile     string
	outputDir        string
	resources        []Resource
	checkBoxes       []widget.Bool
	parser           *mhtmlparser.MHTMLParser
}

func main() {
    go func() {
        window := new(app.Window)
        window.Option(
            app.Title("MHTML File Extractor"),
            app.Size(unit.Dp(800), unit.Dp(600)),
        )
        if err := run(window); err != nil {
            log.Fatal(err)
        }
        os.Exit(0)
    }()
    app.Main()
}

func run(window *app.Window) error {
	theme := material.NewTheme()
	mhtmlApp := &MHTMLApp{
		window:           window,
		theme:            theme,
		darkMode:         true,
		filePath:         widget.Editor{ReadOnly: true},
		rawContent:       widget.Editor{ReadOnly: true},
		resources:        []Resource{},
		checkBoxes:       []widget.Bool{},
		parser:           mhtmlparser.New("", false),
		fetchExternalBtn: widget.Bool{Value: true},
	}
	mhtmlApp.setDarkModePalette()

	var ops op.Ops
    for {
        e := window.Event()
        switch evt := e.(type) {
        case app.DestroyEvent:
            return evt.Err
        case app.FrameEvent:
            gtx := app.NewContext(&ops, evt)
            mhtmlApp.Layout(gtx)
            evt.Frame(gtx.Ops)
        }
    }
}

func (a *MHTMLApp) setDarkModePalette() {
	a.theme.Palette = material.Palette{
		Bg:         color.NRGBA{R: 30, G: 30, B: 30, A: 255},
		Fg:         color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		ContrastBg: color.NRGBA{R: 50, G: 50, B: 50, A: 255},
		ContrastFg: color.NRGBA{R: 30, G: 255, B: 255, A: 255},
	}
}

func (a *MHTMLApp) toggleDarkMode() {
	a.darkMode = !a.darkMode
	if a.darkMode {
		a.setDarkModePalette()
	} else {
		a.theme.Palette = material.Palette{
            Bg:         color.NRGBA{R: 255, G: 255, B: 255, A: 255},
            Fg:         color.NRGBA{R: 30, G: 30, B: 30, A: 255},
            ContrastBg: color.NRGBA{R: 33, G: 150, B: 243, A: 255},
            ContrastFg: color.NRGBA{R: 30, G: 30, B: 30, A: 255},
		}
	}
	a.status = fmt.Sprintf("Theme switched to %s mode", map[bool]string{true: "dark", false: "light"}[a.darkMode])
	a.window.Invalidate()
}

func (a *MHTMLApp) Layout(gtx C) D {
	// Paint the background with theme.Palette.Bg
	defer clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops).Pop()
	paint.ColorOp{Color: a.theme.Palette.Bg}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.UniformInset(unit.Dp(12)).Layout(gtx, func(gtx C) D {
				title := material.H4(a.theme, "🗃 MHTML File Extractor")
				title.Alignment = text.Middle
				return title.Layout(gtx)
			})
		}),
		layout.Rigid(func(gtx C) D {
			return layout.UniformInset(unit.Dp(12)).Layout(gtx, a.fileSelectionRow)
		}),
		layout.Flexed(1, func(gtx C) D {
			return layout.UniformInset(unit.Dp(12)).Layout(gtx, a.rawView)
		}),
		layout.Rigid(func(gtx C) D {
			return layout.UniformInset(unit.Dp(12)).Layout(gtx, func(gtx C) D {
				return material.H6(a.theme, "📦 Embedded Resources").Layout(gtx)
			})
		}),
		layout.Flexed(1, func(gtx C) D {
			return layout.UniformInset(unit.Dp(12)).Layout(gtx, a.resourcesTable)
		}),
		layout.Rigid(func(gtx C) D {
			return layout.UniformInset(unit.Dp(12)).Layout(gtx, a.actionButtons)
		}),
		layout.Rigid(func(gtx C) D {
			return layout.UniformInset(unit.Dp(12)).Layout(gtx, func(gtx C) D {
				return material.Body2(a.theme, "Status: "+a.status).Layout(gtx)
			})
		}),
	)
}
func (a *MHTMLApp) fileSelectionRow(gtx C) D {
	return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceStart}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			for a.browseBtn.Clicked(gtx) {
				a.browseFile()
			}
			return material.Button(a.theme, &a.browseBtn, "📂 Browse").Layout(gtx)
		}),
		layout.Flexed(1, func(gtx C) D {
			return layout.Inset{Left: unit.Dp(8)}.Layout(gtx, func(gtx C) D {
				label := a.filePath.Text()
				if label == "" {
					label = "[No file selected]"
				}
				return material.Body1(a.theme, "Selected File: "+label).Layout(gtx)
			})
		}),
		layout.Rigid(func(gtx C) D {
			for a.fetchExternalBtn.Update(gtx) {
				a.reparseFile()
			}
			return layout.Inset{Left: unit.Dp(8)}.Layout(gtx, func(gtx C) D {
				return material.CheckBox(a.theme, &a.fetchExternalBtn, "Fetch External Scripts").Layout(gtx)
			})
		}),
		layout.Rigid(func(gtx C) D {
			for a.darkModeBtn.Clicked(gtx) {
				a.toggleDarkMode()
			}
			return layout.Inset{Left: unit.Dp(8)}.Layout(gtx, func(gtx C) D {
				return material.Button(a.theme, &a.darkModeBtn, "🌓 Mode").Layout(gtx)
			})
		}),
	)
}

func (a *MHTMLApp) rawView(gtx C) D {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.UniformInset(unit.Dp(8)).Layout(gtx, func(gtx C) D {
				return material.H6(a.theme, "📄 Raw Source").Layout(gtx)
			})
		}),
		layout.Flexed(1, func(gtx C) D {
			return layout.Inset{Top: unit.Dp(8)}.Layout(gtx, func(gtx C) D {
				if a.rawContent.Text() == "" {
					return material.Label(a.theme, unit.Sp(16), "[No content loaded]").Layout(gtx)
				}
				return material.Editor(a.theme, &a.rawContent, "").Layout(gtx)
			})
		}),
	)
}

func (a *MHTMLApp) resourcesTable(gtx C) D {
	list := &layout.List{Axis: layout.Vertical}
	return list.Layout(gtx, len(a.resources)+1, func(gtx C, i int) D {
		if i == 0 {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return layout.Inset{Left: unit.Dp(8)}.Layout(gtx, func(gtx C) D {
						return material.Label(a.theme, unit.Sp(14), "Select").Layout(gtx)
					})
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{Left: unit.Dp(16)}.Layout(gtx, func(gtx C) D {
						return material.Label(a.theme, unit.Sp(14), "Type").Layout(gtx)
					})
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{Left: unit.Dp(16)}.Layout(gtx, func(gtx C) D {
						return material.Label(a.theme, unit.Sp(14), "File Name").Layout(gtx)
					})
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{Left: unit.Dp(16)}.Layout(gtx, func(gtx C) D {
						return material.Label(a.theme, unit.Sp(14), "Size").Layout(gtx)
					})
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{Left: unit.Dp(16)}.Layout(gtx, func(gtx C) D {
						return material.Label(a.theme, unit.Sp(14), "Source").Layout(gtx)
					})
				}),
			)
		}
		res := a.resources[i-1]
		return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				if i-1 >= len(a.checkBoxes) {
					a.checkBoxes = append(a.checkBoxes, widget.Bool{Value: res.Selected})
				}
				chk := &a.checkBoxes[i-1]
				if chk.Value != res.Selected {
					res.Selected = chk.Value
					a.resources[i-1].Selected = chk.Value
				}
				return material.CheckBox(a.theme, chk, "").Layout(gtx)
			}),
			layout.Rigid(func(gtx C) D {
				return layout.Inset{Left: unit.Dp(16)}.Layout(gtx, func(gtx C) D {
					return material.Label(a.theme, unit.Sp(14), res.Type).Layout(gtx)
				})
			}),
			layout.Rigid(func(gtx C) D {
				return layout.Inset{Left: unit.Dp(16)}.Layout(gtx, func(gtx C) D {
					return material.Label(a.theme, unit.Sp(14), res.Filename).Layout(gtx)
				})
			}),
			layout.Rigid(func(gtx C) D {
				return layout.Inset{Left: unit.Dp(16)}.Layout(gtx, func(gtx C) D {
					return material.Label(a.theme, unit.Sp(14), formatSize(res.Size)).Layout(gtx)
				})
			}),
			layout.Rigid(func(gtx C) D {
				return layout.Inset{Left: unit.Dp(16)}.Layout(gtx, func(gtx C) D {
					return material.Label(a.theme, unit.Sp(14), res.Source).Layout(gtx)
				})
			}),
		)
	})
}

func (a *MHTMLApp) actionButtons(gtx C) D {
	return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceBetween}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			for a.selectAllBtn.Clicked(gtx) {
				a.selectAllResources()
			}
			return material.Button(a.theme, &a.selectAllBtn, "✓ Select All").Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			for a.extractBtn.Clicked(gtx) {
				a.extractSelected()
			}
			return material.Button(a.theme, &a.extractBtn, "⬇️ Extract Selected").Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			for a.outputDirBtn.Clicked(gtx) {
				a.changeOutputDir()
			}
			return material.Button(a.theme, &a.outputDirBtn, "🗂 Change Output Dir").Layout(gtx)
		}),
	)
}

func (a *MHTMLApp) browseFile() {
	filePath, err := zenity.SelectFile(
		zenity.Title("Select MHTML File"),
		zenity.FileFilters{
			{Name: "MHTML Files", Patterns: []string{"*.mhtml", "*.mht"}, CaseFold: true},
		},
	)
	if err == zenity.ErrCanceled {
		a.status = "File selection canceled"
		a.window.Invalidate()
		return
	}
	if err != nil {
		a.status = fmt.Sprintf("Error selecting file: %v", err)
		a.window.Invalidate()
		return
	}
	if filePath != "" {
		a.selectedFile = filePath
		a.filePath.SetText(filePath)
		baseName := filepath.Base(filePath)
		ext := filepath.Ext(baseName)
		baseName = baseName[:len(baseName)-len(ext)]
		a.outputDir = filepath.Join(filepath.Dir(filePath), baseName)
		a.parseMHTML()
	}
}

func (a *MHTMLApp) reparseFile() {
	if a.selectedFile != "" {
		a.parseMHTML()
	}
}

func (a *MHTMLApp) parseMHTML() {
	a.parser = mhtmlparser.New(a.selectedFile, a.fetchExternalBtn.Value)
	if err := a.parser.Parse(); err != nil {
		a.status = fmt.Sprintf("Error parsing MHTML file: %v", err)
		a.window.Invalidate()
		return
	}

	// Populate resources
	a.resources = make([]Resource, len(a.parser.Resources))
	a.checkBoxes = make([]widget.Bool, len(a.parser.Resources))
	for i, res := range a.parser.Resources {
		a.resources[i] = Resource{
			Type:     res.Type,
			Filename: res.Filename,
			Size:     int64(res.Size),
			Source:   res.Source,
			Selected: true,
		}
		a.checkBoxes[i] = widget.Bool{Value: true}
	}

	htmlContent := a.parser.GetHTMLContent()
	if htmlContent == "" {
		htmlContent = "[No HTML content found]"
	}
	a.rawContent.SetText(htmlContent)
	a.status = fmt.Sprintf("Loaded: %s (Output directory: %s, %d resources found)", a.selectedFile, a.outputDir, len(a.resources))
	a.window.Invalidate()
}

func (a *MHTMLApp) changeOutputDir() {
	dirPath, err := zenity.SelectFile(
		zenity.Title("Select Output Directory"),
		zenity.Directory(),
	)
	if err == zenity.ErrCanceled {
		a.status = "Directory selection canceled"
		a.window.Invalidate()
		return
	}
	if err != nil {
		a.status = fmt.Sprintf("Error selecting directory: %v", err)
		a.window.Invalidate()
		return
	}
	if dirPath != "" {
		a.outputDir = dirPath
		a.status = "Output directory set to: " + a.outputDir
		a.window.Invalidate()
	}
}

func (a *MHTMLApp) selectAllResources() {
	selectAll := true
	if len(a.resources) > 0 && a.resources[0].Selected {
		selectAll = false
	}
	for i := range a.resources {
		a.resources[i].Selected = selectAll
		if i < len(a.checkBoxes) {
			a.checkBoxes[i].Value = selectAll
		}
	}
	a.status = "All resources selected"
	if !selectAll {
		a.status = "All resources deselected"
	}
	a.window.Invalidate()
}

func (a *MHTMLApp) extractSelected() {
	if a.selectedFile == "" {
		a.status = "No MHTML file selected"
		a.window.Invalidate()
		return
	}
	if a.outputDir == "" {
		a.status = "No output directory set"
		a.window.Invalidate()
		return
	}

	selectedIndices := []int{}
	for i, res := range a.resources {
		if res.Selected {
			selectedIndices = append(selectedIndices, i)
		}
	}

	paths, err := a.parser.ExtractResources(a.outputDir, selectedIndices)
	if err != nil {
		a.status = fmt.Sprintf("Error extracting resources: %v", err)
		a.window.Invalidate()
		return
	}

	if len(paths) == 0 {
		a.status = "No resources selected for extraction"
	} else {
		a.status = fmt.Sprintf("Extracted %d resources to %s", len(paths), a.outputDir)
	}
	a.window.Invalidate()
}

func formatSize(size int64) string {
	return fmt.Sprintf("%.2f KB", float64(size)/1024)
}