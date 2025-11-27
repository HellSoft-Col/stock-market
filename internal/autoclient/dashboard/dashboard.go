package dashboard

import (
	"fmt"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TraderMetrics holds real-time metrics for a single trader
type TraderMetrics struct {
	TeamName        string
	Strategy        string
	Balance         float64
	InventoryValue  float64
	NetWorth        float64
	PnL             float64
	PnLPercent      float64
	OrdersPlaced    int
	FillsReceived   int
	ActiveOrders    int
	LastAction      string
	LastActionTime  time.Time
	ErrorCount      int
	Status          string // "connecting", "connected", "active", "degraded", "error", "waiting"
	AIDecisions     int
	ProductionCount int
	Connected       bool
	Authenticated   bool
}

// Dashboard is the main TUI model
type Dashboard struct {
	traders     map[string]*TraderMetrics
	mu          sync.RWMutex
	width       int
	height      int
	startTime   time.Time
	globalPnL   float64
	totalOrders int
	totalFills  int
	lastUpdate  time.Time
}

// NewDashboard creates a new dashboard instance
func NewDashboard() *Dashboard {
	return &Dashboard{
		traders:    make(map[string]*TraderMetrics),
		startTime:  time.Now(),
		lastUpdate: time.Now(),
	}
}

// UpdateTrader updates metrics for a specific trader
func (d *Dashboard) UpdateTrader(teamName string, metrics *TraderMetrics) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.traders[teamName] = metrics
	d.lastUpdate = time.Now()
}

// GetTrader retrieves metrics for a specific trader
func (d *Dashboard) GetTrader(teamName string) *TraderMetrics {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.traders[teamName]
}

// Init initializes the dashboard
func (d *Dashboard) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		tickCmd(),
	)
}

// Update handles messages and updates the model
func (d *Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.width = msg.Width
		d.height = msg.Height
		return d, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return d, tea.Quit
		}

	case tickMsg:
		return d, tickCmd()
	}

	return d, nil
}

// View renders the dashboard
func (d *Dashboard) View() string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.width == 0 {
		return "Loading..."
	}

	// Styles
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		Background(lipgloss.Color("235")).
		Padding(0, 2).
		Width(d.width)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Align(lipgloss.Center)

	// Header
	uptime := time.Since(d.startTime).Round(time.Second)
	header := headerStyle.Render(fmt.Sprintf(
		"🚀 Stock Market AI Trading Dashboard | Uptime: %s | Traders: %d | Last Update: %s",
		uptime, len(d.traders), d.lastUpdate.Format("15:04:05"),
	))

	// Title
	title := titleStyle.Width(d.width).Render("═══ AI TRADERS IN ACTION ═══")

	// Calculate grid layout (3 columns for 11 traders)
	panelWidth := (d.width - 6) / 3
	panelHeight := 12

	// Create trader panels
	var panels []string
	var currentRow []string
	idx := 0

	// Sort traders by name for consistent display
	traderNames := []string{
		"Alquimistas de Palta",
		"Arpistas de Pita-Pita",
		"Avocultores del Hueso Cósmico",
		"Cartógrafos de Fosfolima",
		"Cosechadores de Semillas",
		"Forjadores Holográficos",
		"Ingenieros Holo-Aguacate",
		"Mensajeros del Núcleo",
		"Monjes del Guacamole Estelar",
		"Orfebres de Cáscara",
		"Someliers de Aceite",
	}

	for _, name := range traderNames {
		trader := d.traders[name]
		if trader == nil {
			// Create empty panel
			trader = &TraderMetrics{
				TeamName: name,
				Status:   "waiting",
			}
		}

		panel := d.renderTraderPanel(trader, panelWidth, panelHeight)
		currentRow = append(currentRow, panel)
		idx++

		// 3 panels per row
		if idx%3 == 0 {
			panels = append(panels, lipgloss.JoinHorizontal(lipgloss.Top, currentRow...))
			currentRow = []string{}
		}
	}

	// Add remaining panels
	if len(currentRow) > 0 {
		panels = append(panels, lipgloss.JoinHorizontal(lipgloss.Top, currentRow...))
	}

	// Join all rows
	grid := lipgloss.JoinVertical(lipgloss.Left, panels...)

	// Footer with global stats
	d.calculateGlobalStats()
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Background(lipgloss.Color("235")).
		Padding(0, 2).
		Width(d.width)

	footer := footerStyle.Render(fmt.Sprintf(
		"Global P&L: %.2f%% | Total Orders: %d | Total Fills: %d | Press 'q' to quit",
		d.globalPnL, d.totalOrders, d.totalFills,
	))

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		title,
		"",
		grid,
		"",
		footer,
	)
}

// renderTraderPanel renders a single trader panel
func (d *Dashboard) renderTraderPanel(trader *TraderMetrics, width, height int) string {
	// Status indicator with detailed connection state
	statusEmoji := "⏸️"
	statusColor := lipgloss.Color("240")
	statusText := ""

	switch trader.Status {
	case "connecting":
		statusEmoji = "🔄"
		statusColor = lipgloss.Color("33")
		statusText = "CONN..."
	case "connected":
		statusEmoji = "🔐"
		statusColor = lipgloss.Color("226")
		statusText = "AUTH..."
	case "active":
		statusEmoji = "✅"
		statusColor = lipgloss.Color("46")
		if trader.Connected && trader.Authenticated {
			statusText = "LIVE"
		} else {
			statusText = "READY"
		}
	case "degraded":
		statusEmoji = "⚠️"
		statusColor = lipgloss.Color("226")
		statusText = "SLOW"
	case "error":
		statusEmoji = "❌"
		statusColor = lipgloss.Color("196")
		statusText = "ERROR"
	case "waiting":
		statusEmoji = "⏳"
		statusColor = lipgloss.Color("240")
		statusText = "WAIT"
	default:
		statusEmoji = "⏸️"
		statusColor = lipgloss.Color("240")
		statusText = "IDLE"
	}

	// P&L indicator
	pnlEmoji := "📊"
	if trader.PnLPercent > 0 {
		pnlEmoji = "📈"
	} else if trader.PnLPercent < 0 {
		pnlEmoji = "📉"
	}

	// Personality icon based on team name
	personality := getPersonalityIcon(trader.TeamName)

	// Panel border style
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(statusColor).
		Padding(0, 1).
		Width(width).
		Height(height)

	// Team name (shortened)
	teamShort := shortenTeamName(trader.TeamName)
	teamStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39"))

	// Content with status text
	statusStyle := lipgloss.NewStyle().
		Foreground(statusColor).
		Bold(true)

	content := fmt.Sprintf("%s %s %s %s\n",
		statusEmoji,
		teamStyle.Render(teamShort),
		personality,
		statusStyle.Render(statusText),
	)
	content += lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render(strings.Repeat("─", width-4)) + "\n"

	// Balance and P&L
	content += fmt.Sprintf("💰 $%.0f | %s %.1f%%\n", trader.NetWorth, pnlEmoji, trader.PnLPercent)

	// Activity bars
	orderBar := createProgressBar(trader.OrdersPlaced, 100, 15, lipgloss.Color("33"))
	fillBar := createProgressBar(trader.FillsReceived, 100, 15, lipgloss.Color("213"))
	aiBar := createProgressBar(trader.AIDecisions, 50, 15, lipgloss.Color("170"))

	content += fmt.Sprintf("📤 %s %d\n", orderBar, trader.OrdersPlaced)
	content += fmt.Sprintf("📥 %s %d\n", fillBar, trader.FillsReceived)
	content += fmt.Sprintf("🤖 %s %d\n", aiBar, trader.AIDecisions)

	// Active orders
	if trader.ActiveOrders > 0 {
		content += lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Render(
			fmt.Sprintf("⏳ %d active\n", trader.ActiveOrders),
		)
	} else {
		content += lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("⏳ 0 active\n")
	}

	// Last action
	if trader.LastAction != "" {
		timeSince := time.Since(trader.LastActionTime)
		if timeSince < 5*time.Second {
			content += lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Render(
				fmt.Sprintf("⚡ %s", truncate(trader.LastAction, 20)),
			)
		} else {
			content += lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(
				fmt.Sprintf("💤 %ds ago", int(timeSince.Seconds())),
			)
		}
	}

	return borderStyle.Render(content)
}

// Helper functions

func getPersonalityIcon(teamName string) string {
	personalities := map[string]string{
		"Alquimistas de Palta":          "🏭",  // Bill Gates - Production
		"Arpistas de Pita-Pita":         "🚀",  // Elon Musk - Disruptor
		"Avocultores del Hueso Cósmico": "🦈",  // Carl Icahn - Aggressive
		"Cartógrafos de Fosfolima":      "🎯",  // Peter Thiel - Contrarian
		"Cosechadores de Semillas":      "📊",  // Cathie Wood - Growth
		"Forjadores Holográficos":       "⚖️", // Ray Dalio - Balanced
		"Ingenieros Holo-Aguacate":      "🔍",  // Michael Burry - Value
		"Mensajeros del Núcleo":         "🤖",  // AI Traders
		"Monjes del Guacamole Estelar":  "💥",  // CHAOS AGENT
		"Orfebres de Cáscara":           "⚙️", // Production Beast
		"Someliers de Aceite":           "💸",  // Sales Beast
	}

	if icon, ok := personalities[teamName]; ok {
		return icon
	}
	return "🎲"
}

func shortenTeamName(name string) string {
	shorts := map[string]string{
		"Alquimistas de Palta":          "PALTA",
		"Arpistas de Pita-Pita":         "PITA",
		"Avocultores del Hueso Cósmico": "HUESO",
		"Cartógrafos de Fosfolima":      "FOSFO",
		"Cosechadores de Semillas":      "SEMILLAS",
		"Forjadores Holográficos":       "HOLO",
		"Ingenieros Holo-Aguacate":      "AGUACATE",
		"Mensajeros del Núcleo":         "NÚCLEO",
		"Monjes del Guacamole Estelar":  "CHAOS",
		"Orfebres de Cáscara":           "CÁSCARA",
		"Someliers de Aceite":           "ACEITE",
	}

	if short, ok := shorts[name]; ok {
		return short
	}
	return name
}

func createProgressBar(value, max, width int, color lipgloss.Color) string {
	if max == 0 {
		max = 1
	}

	percentage := float64(value) / float64(max)
	if percentage > 1.0 {
		percentage = 1.0
	}

	filled := int(float64(width) * percentage)
	empty := width - filled

	barStyle := lipgloss.NewStyle().Foreground(color)
	emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("235"))

	bar := barStyle.Render(strings.Repeat("█", filled))
	bar += emptyStyle.Render(strings.Repeat("░", empty))

	return bar
}

func truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length-3] + "..."
}

func (d *Dashboard) calculateGlobalStats() {
	totalPnL := 0.0
	count := 0
	d.totalOrders = 0
	d.totalFills = 0

	for _, trader := range d.traders {
		if trader.Status != "waiting" {
			totalPnL += trader.PnLPercent
			count++
			d.totalOrders += trader.OrdersPlaced
			d.totalFills += trader.FillsReceived
		}
	}

	if count > 0 {
		d.globalPnL = totalPnL / float64(count)
	}
}

// Tick message for periodic updates
type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
