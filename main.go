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

// GameState tracks whether we're on the title screen or in the game
type GameState int

const (
	TitleScreen GameState = iota
	Playing
	GameOver
	Win
)

// World contains the game state
type World struct {
	Player   Player
	Width    int
	Height   int
	Terrain  [][]int
	GameOver bool
	Score    int
}

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
			Score: 0,
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

	// Place the player in the river at the bottom
	// Find a point in the last row that is part of the river
	lastRow := height - 1

	// Find the leftmost and rightmost points of the river in the last row
	leftmost := -1
	rightmost := -1
	for x := 0; x < width; x++ {
		if terrain[lastRow][x] == River {
			if leftmost == -1 {
				leftmost = x
			}
			rightmost = x
		}
	}

	// Place the player in the middle of the river
	if leftmost != -1 && rightmost != -1 {
		m.world.Player.X = leftmost + (rightmost-leftmost)/2
		m.world.Player.Y = lastRow
	} else {
		// Fallback if no river is found
		for x := 0; x < width; x++ {
			if terrain[lastRow][x] == River {
				m.world.Player.X = x
				m.world.Player.Y = lastRow
				break
			}
		}
	}

	return m
}

// Check if the player has collided with land
func (m model) checkCollision() model {
	playerX := m.world.Player.X
	playerY := m.world.Player.Y

	// Check boundaries
	if playerY < 0 || playerY >= len(m.world.Terrain) ||
		playerX < 0 || playerX >= len(m.world.Terrain[0]) {
		m.state = GameOver
		return m
	}

	// Check if player is on land
	if m.world.Terrain[playerY][playerX] == Land {
		m.state = GameOver
		return m
	}

	// Check for win condition - reaching the top row
	if playerY == 0 {
		m.state = Win
		return m
	}

	// If we're still playing, increment score
	// Only increment score when moving up (away from starting position)
	if m.world.Player.Y < len(m.world.Terrain)-1 {
		m.world.Score++
	}

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

		case "s", "r":
			// Start game with 's' or restart with 'r'
			if m.state == TitleScreen || m.state == GameOver || m.state == Win {
				m.state = Playing
				m.world.Score = 0
				// Initialize the terrain when starting the game
				m = m.initializeTerrain()
			}

		// Add player movement controls when in playing state
		case "up", "k":
			if m.state == Playing && m.world.Player.Y > 0 {
				m.world.Player.Y--
				m = m.checkCollision()
			}

		case "down", "j":
			if m.state == Playing && m.world.Player.Y < m.height-3 {
				m.world.Player.Y++
				m = m.checkCollision()
			}

		case "left", "h":
			if m.state == Playing && m.world.Player.X > 0 {
				m.world.Player.X--
				m = m.checkCollision()
			}

		case "right", "l":
			if m.state == Playing && m.world.Player.X < m.width-1 {
				m.world.Player.X++
				m = m.checkCollision()
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update the world dimensions
		m.world.Width = msg.Width
		m.world.Height = msg.Height

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

	// Add score and instructions to status line
	view += fmt.Sprintf("\nScore: %d | Navigate to the top! Use arrow keys to move. Avoid land (.) | Press q to quit", m.world.Score)

	return view
}

func (m model) renderGameOverScreen() string {
	// Create a stylish game over message
	gameOverStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF0000")).
		Background(lipgloss.Color("#000000")).
		PaddingLeft(2).
		PaddingRight(2)

	gameOver := gameOverStyle.Render("GAME OVER")
	gameOver = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, gameOver)

	// Create score message
	scoreStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF"))

	scoreMsg := fmt.Sprintf("Final Score: %d", m.world.Score)
	scoreMsg = scoreStyle.Render(scoreMsg)
	scoreMsg = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, scoreMsg)

	// Create restart instruction
	restartMsg := "Press 'r' to restart"
	restartMsg = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, restartMsg)

	// Center vertically
	verticalPadding := m.height/2 - 3
	view := ""
	for i := 0; i < verticalPadding; i++ {
		view += "\n"
	}

	view += gameOver + "\n\n"
	view += scoreMsg + "\n\n"
	view += restartMsg + "\n\n"

	// Add quit message at bottom
	view += lipgloss.PlaceHorizontal(m.width, lipgloss.Center, "Press q to quit")

	return view
}

func (m model) renderWinScreen() string {
	// Create a stylish win message
	winStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFF00")).
		Background(lipgloss.Color("#009900")).
		PaddingLeft(2).
		PaddingRight(2)

	winMsg := winStyle.Render("YOU WIN!")
	winMsg = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, winMsg)

	// Create score message
	scoreStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF"))

	scoreMsg := fmt.Sprintf("Final Score: %d", m.world.Score)
	scoreMsg = scoreStyle.Render(scoreMsg)
	scoreMsg = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, scoreMsg)

	// Create restart instruction
	restartMsg := "Press 'r' to play again"
	restartMsg = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, restartMsg)

	// Center vertically
	verticalPadding := m.height/2 - 3
	view := ""
	for i := 0; i < verticalPadding; i++ {
		view += "\n"
	}

	view += winMsg + "\n\n"
	view += scoreMsg + "\n\n"
	view += restartMsg + "\n\n"

	// Add quit message at bottom
	view += lipgloss.PlaceHorizontal(m.width, lipgloss.Center, "Press q to quit")

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
	case GameOver:
		return m.renderGameOverScreen()
	case Win:
		return m.renderWinScreen()
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
