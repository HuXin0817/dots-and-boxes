package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

var (
	// CurrentThemeVariant Global variables for current gameTheme variant and game state
	CurrentThemeVariant fyne.ThemeVariant

	LightThemeDotCanvasColor = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF} // #FFFFFF
	DarkThemeDotCanvasColor  = color.NRGBA{R: 0xCA, G: 0xCA, B: 0xCA, A: 0xFF} // #CACACA
	LightThemeColor          = color.NRGBA{R: 0xF2, G: 0xF2, B: 0xF2, A: 0xFF} // #F2F2F2
	DarkThemeColor           = color.NRGBA{R: 0x2B, G: 0x2B, B: 0x2B, A: 0xFF} // #2B2B2B
	LightThemeButtonColor    = color.NRGBA{R: 0xD9, G: 0xD9, B: 0xD9, A: 0xFF} // #D9D9D9
	DarkThemeButtonColor     = color.NRGBA{R: 0x41, G: 0x41, B: 0x41, A: 0xFF} // #414141
	Player1HighlightColor    = color.NRGBA{R: 0x40, G: 0x40, B: 0xFF, A: 0x80} // #4040FF80
	Player2HighlightColor    = color.NRGBA{R: 0xFF, G: 0x40, B: 0x40, A: 0x80} // #FF404080
	Player1FilledColor       = color.NRGBA{R: 0x40, G: 0x40, B: 0xFF, A: 0x40} // #4040FF40
	Player2FilledColor       = color.NRGBA{R: 0xFF, G: 0x40, B: 0x40, A: 0x40} // #FF404040
	TipColor                 = color.NRGBA{R: 0xFF, G: 0xFF, B: 0x40, A: 0x40} // #FFFF4040
)

// GameTheme implements the fyne.Theme interface
type GameTheme struct{}

var gameTheme = &GameTheme{}

// Color returns the color for a given gameTheme element and variant (light/dark)
func (g *GameTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// If the gameTheme variant has changed, update the colors of the canvas elements
	if CurrentThemeVariant != variant {
		CurrentThemeVariant = variant
		// Update the color of dot canvases
		for _, circle := range DotCanvases {
			circle.FillColor = g.GetDotCanvasColor()
			circle.Refresh()
		}
		// Update the color of box canvases
		boxesCanvasLock.Lock()
		for box, rectangle := range BoxesCanvases {
			if _, c := BoxesFilledColor[box]; !c {
				rectangle.FillColor = g.GetThemeColor()
				rectangle.Refresh()
			}
		}
		boxesCanvasLock.Unlock()
	}

	// Return the appropriate color based on the element name
	switch name {
	case theme.ColorNameBackground:
		return g.GetThemeColor()
	case theme.ColorNameButton:
		return g.GetButtonColor()
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

// Icon returns the main window icon for the given gameTheme icon name
func (g *GameTheme) Icon(fyne.ThemeIconName) fyne.Resource { return MainWindow.Icon() }

// Font returns the font resource for the given text style
func (g *GameTheme) Font(style fyne.TextStyle) fyne.Resource { return theme.DefaultTheme().Font(style) }

// Size returns the size for the given gameTheme size name
func (g *GameTheme) Size(name fyne.ThemeSizeName) float32 { return theme.DefaultTheme().Size(name) }

// getColorByVariant returns the appropriate color based on the current gameTheme variant
func (g *GameTheme) getColorByVariant(lightColor, darkColor color.Color) color.Color {
	if CurrentThemeVariant == theme.VariantDark {
		return darkColor
	} else {
		return lightColor
	}
}

// GetDotCanvasColor returns the color for dot canvases based on the current gameTheme variant
func (g *GameTheme) GetDotCanvasColor() color.Color {
	return g.getColorByVariant(LightThemeDotCanvasColor, DarkThemeDotCanvasColor)
}

// GetThemeColor returns the general gameTheme color based on the current gameTheme variant
func (g *GameTheme) GetThemeColor() color.Color {
	return g.getColorByVariant(LightThemeColor, DarkThemeColor)
}

// GetButtonColor returns the button color based on the current gameTheme variant
func (g *GameTheme) GetButtonColor() color.Color {
	return g.getColorByVariant(LightThemeButtonColor, DarkThemeButtonColor)
}

// GetPlayerFilledColor returns the color used to fill boxes based on the current player's turn
func (g *GameTheme) GetPlayerFilledColor() color.Color {
	if CurrentTurn == Player1Turn {
		return Player1FilledColor
	} else {
		return Player2FilledColor
	}
}

// GetPlayerHighlightColor returns the color used to highlight moves based on the current player's turn
func (g *GameTheme) GetPlayerHighlightColor() color.Color {
	if CurrentTurn == Player1Turn {
		return Player1HighlightColor
	} else {
		return Player2HighlightColor
	}
}
