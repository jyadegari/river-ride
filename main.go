package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Player represents the character in the game
type Player struct {
	X int
	Y int
}

// Terrain types
const (
	Land = iota
	River
)

// World contains the game state
type World struct {
	Player  Player
	Width   int
	Height  int
	Terrain [][]int
}

// GameState tracks whether we're on the title screen or in the game
type GameState int

const (
	TitleScreen GameState = iota
	Playing
)

type model struct {
	state  GameState
	world  World
	width  int
	height int
}

func initialModel() model {
	// Start on the title screen
	return model{
		state: TitleScreen,
		world: World{
			Player: Player{
				X: 10,
				Y: 5,
			},
		},
	}
}

// Initialize the world terrain with a river
func (m model) initializeTerrain() model {
	width := m.width
	height := m.height - 2 // Account for status area

	// Create terrain grid
	terrain := make([][]int, height)
	for y := 0; y < height; y++ {
		terrain[y] = make([]int, width)
		// Initialize everything as land
		for x := 0; x < width; x++ {
			terrain[y][x] = Land
		}
	}

	// Create a winding river through the middle
	// Random river width between 10 and 20
	riverWidth := rand.Intn(11) + 10
	centerX := width / 2

	// Generate a winding river
	for y := 0; y < height; y++ {
		// Make the river wind using a sine wave
		offset := int(math.Sin(float64(y)/5) * 10)
		riverCenter := centerX + offset

		// Create the river with specified width
		for x := riverCenter - riverWidth/2; x <= riverCenter+riverWidth/2; x++ {
			if x >= 0 && x < width {
				terrain[y][x] = River
			}
		}

		// Occasionally add small river branches
		if rand.Intn(20) == 0 {
			branchLength := rand.Intn(10) + 5
			branchDirection := 1
			if rand.Intn(2) == 0 {
				branchDirection = -1
			}

			for i := 0; i < branchLength; i++ {
				branchX := riverCenter + (i+1)*branchDirection
				if branchX >= 0 && branchX < width && y+i < height {
					terrain[y+i][branchX] = River
				}
			}
		}
	}

	m.world.Terrain = terrain
	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "s":
			if m.state == TitleScreen {
				m.state = Playing
				// Initialize the terrain when starting the game
				m = m.initializeTerrain()
			}

		// Add player movement controls when in playing state
		case "up", "k":
			if m.state == Playing && m.world.Player.Y > 0 {
				m.world.Player.Y--
			}

		case "down", "j":
			if m.state == Playing && m.world.Player.Y < m.height-3 {
				m.world.Player.Y++
			}

		case "left", "h":
			if m.state == Playing && m.world.Player.X > 0 {
				m.world.Player.X--
			}

		case "right", "l":
			if m.state == Playing && m.world.Player.X < m.width-1 {
				m.world.Player.X++
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update the world dimensions
		m.world.Width = msg.Width
		m.world.Height = msg.Height

		// Center the player when window size changes
		m.world.Player.X = msg.Width / 2
		m.world.Player.Y = msg.Height / 2

		// Re-initialize terrain if playing
		if m.state == Playing {
			m = m.initializeTerrain()
		}
	}

	return m, nil
}

func (m model) renderTitleScreen() string {
	// Create the title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		PaddingLeft(2).
		PaddingRight(2)

	title := titleStyle.Render("River Ride")
	title = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, title)

	// Center vertically with empty lines
	verticalPadding := m.height/2 - 2
	view := ""
	for i := 0; i < verticalPadding; i++ {
		view += "\n"
	}

	view += title + "\n\n"

	// Add instructions
	instructions := "Press 's' to start the game"
	instructions = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, instructions)
	view += instructions + "\n"

	// Add quit message at bottom left
	// Add empty lines to push the quit message to the bottom
	remainingLines := m.height - verticalPadding - 5 // Adjust for title and instructions
	for i := 0; i <= remainingLines; i++ {
		view += "\n"
	}

	view += "Press q to quit"

	return view
}

func (m model) renderGameScreen() string {
	// Create a grid for the game world
	grid := make([][]string, m.height)
	for y := 0; y < m.height; y++ {
		grid[y] = make([]string, m.width)
		for x := 0; x < m.width; x++ {
			grid[y][x] = " "
		}
	}

	// Render the terrain
	terrainHeight := len(m.world.Terrain)
	if terrainHeight > 0 {
		terrainWidth := len(m.world.Terrain[0])

		for y := 0; y < terrainHeight && y < m.height-2; y++ {
			for x := 0; x < terrainWidth && x < m.width; x++ {
				if m.world.Terrain[y][x] == Land {
					grid[y][x] = "."
				} else if m.world.Terrain[y][x] == River {
					grid[y][x] = " "
				}
			}
		}
	}

	// Place the player on the grid
	player := "P"
	if m.world.Player.Y < m.height && m.world.Player.X < m.width {
		grid[m.world.Player.Y][m.world.Player.X] = player
	}

	// Convert the grid to a string
	view := ""
	for y := 0; y < m.height-2; y++ {
		line := ""
		for x := 0; x < m.width; x++ {
			line += grid[y][x]
		}
		view += line + "\n"
	}

	view += "\nUse arrow keys to move. Press q to quit."

	return view
}

func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Render different screens based on game state
	switch m.state {
	case TitleScreen:
		return m.renderTitleScreen()
	case Playing:
		return m.renderGameScreen()
	default:
		return "Unknown state"
	}
}

func main() {

	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}
