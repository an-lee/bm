package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pkg/browser"

	"bm/internal/app"
	"bm/internal/clipboard"
	"bm/internal/config"
	"bm/internal/search"
	"bm/internal/streams"
)

const tabSearch, tabStreams, tabAddons, tabSettings = 0, 1, 2, 3

type searchDoneMsg []search.TitleResult
type searchErrMsg struct{ err error }
type streamsDoneMsg []streams.ResolvedStream
type streamsErrMsg struct{ err error }
type statusMsg struct {
	err  error
	text string
}
type toastClearMsg struct{}

// rootModel is the Bubble Tea root; use pointer receiver so updates persist.
type rootModel struct {
	app *app.App

	width, height int
	tab           int

	searchInput textinput.Model
	searchList  list.Model
	searchItems []list.Item

	streamsList list.Model
	streamItems []list.Item
	selected    *search.TitleResult
	streamsBusy bool

	addonList    list.Model
	addonItems   []list.Item
	addonURLMode bool
	addonURL     textinput.Model

	settingsInput textinput.Model

	toast string
}

func newRootModel(ap *app.App) *rootModel {
	si := textinput.New()
	si.Placeholder = "Search movies & series…"
	si.CharLimit = 200
	si.Width = 50
	si.Focus()

	sl := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	sl.Title = "Results"
	sl.SetShowStatusBar(false)
	sl.DisableQuitKeybindings()

	strl := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	strl.Title = "Streams"
	strl.SetShowStatusBar(false)
	strl.DisableQuitKeybindings()

	al := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	al.Title = "Addons"
	al.SetShowStatusBar(false)
	al.DisableQuitKeybindings()

	aui := textinput.New()
	aui.Placeholder = "https://…/manifest.json"
	aui.CharLimit = 512
	aui.Width = 60

	tmdb := textinput.New()
	tmdb.Placeholder = "TMDB API key (optional)"
	tmdb.EchoMode = textinput.EchoPassword
	tmdb.CharLimit = 128
	tmdb.Width = 40

	m := &rootModel{
		app:           ap,
		searchInput:   si,
		searchList:    sl,
		streamsList:   strl,
		addonList:     al,
		addonURL:      aui,
		settingsInput: tmdb,
	}
	m.refreshAddonList()
	return m
}

func (m *rootModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.tickToast())
}

func (m *rootModel) tickToast() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return toastClearMsg{}
	})
}

func (m *rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		h := max(6, m.height-8)
		m.searchList.SetSize(max(20, m.width-4), h)
		m.streamsList.SetSize(max(20, m.width-4), h)
		m.addonList.SetSize(max(20, m.width-4), h)
		m.searchInput.Width = max(20, m.width-8)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			if m.addonURLMode {
				m.addonURLMode = false
				m.addonURL.Blur()
				return m, nil
			}
			return m, tea.Quit
		case "tab":
			m.tab = (m.tab + 1) % 4
			return m, m.onTabChange()
		case "shift+tab":
			m.tab = (m.tab + 3) % 4
			return m, m.onTabChange()
		case "1":
			m.tab = tabSearch
			return m, m.onTabChange()
		case "2":
			m.tab = tabStreams
			return m, m.onTabChange()
		case "3":
			m.tab = tabAddons
			return m, m.onTabChange()
		case "4":
			m.tab = tabSettings
			return m, m.onTabChange()
		}

		switch m.tab {
		case tabSearch:
			return m.updateSearch(msg)
		case tabStreams:
			return m.updateStreams(msg)
		case tabAddons:
			return m.updateAddons(msg)
		case tabSettings:
			return m.updateSettings(msg)
		}

	case searchDoneMsg:
		m.searchItems = make([]list.Item, 0, len(msg))
		for _, r := range msg {
			m.searchItems = append(m.searchItems, titleItem{r: r})
		}
		m.searchList.SetItems(m.searchItems)
		m.toast = fmt.Sprintf("%d results", len(msg))
		return m, m.tickToast()

	case searchErrMsg:
		m.toast = "search: " + msg.err.Error()
		return m, m.tickToast()

	case streamsDoneMsg:
		m.streamsBusy = false
		m.streamItems = make([]list.Item, 0, len(msg))
		for _, s := range msg {
			m.streamItems = append(m.streamItems, streamItem{s: s})
		}
		m.streamsList.SetItems(m.streamItems)
		m.toast = fmt.Sprintf("%d streams", len(msg))
		return m, m.tickToast()

	case streamsErrMsg:
		m.streamsBusy = false
		m.toast = "streams: " + msg.err.Error()
		return m, m.tickToast()

	case statusMsg:
		m.addonURLMode = false
		m.addonURL.Blur()
		if msg.err != nil {
			m.toast = msg.err.Error()
			return m, m.tickToast()
		}
		m.toast = msg.text
		_ = m.app.Reload()
		m.refreshAddonList()
		return m, m.tickToast()

	case toastClearMsg:
		m.toast = ""
		return m, nil
	}

	return m, nil
}

func (m *rootModel) onTabChange() tea.Cmd {
	if m.tab == tabSearch {
		return textinput.Blink
	}
	if m.tab == tabSettings {
		m.settingsInput.Focus()
		return textinput.Blink
	}
	m.settingsInput.Blur()
	return nil
}

func (m *rootModel) refreshAddonList() {
	items := make([]list.Item, 0, len(m.app.Config.Addons))
	for _, a := range m.app.Config.Addons {
		st := "off"
		if a.Enabled {
			st = "on"
		}
		items = append(items, addonItem{a: a, label: fmt.Sprintf("%s — %s [%s]", a.ID, a.Name, st)})
	}
	m.addonList.SetItems(items)
}

func (m *rootModel) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.searchInput.Focused() {
		switch msg.String() {
		case "enter":
			q := strings.TrimSpace(m.searchInput.Value())
			if q == "" {
				return m, nil
			}
			return m, m.runSearch(q)
		case "down":
			m.searchInput.Blur()
		default:
			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			return m, cmd
		}
	}

	switch msg.String() {
	case "up", "down", "k", "j":
		var cmd tea.Cmd
		m.searchList, cmd = m.searchList.Update(msg)
		return m, cmd
	case "enter":
		if it, ok := m.searchList.SelectedItem().(titleItem); ok {
			m.selected = &it.r
			m.tab = tabStreams
			return m, m.loadStreamsForSelection()
		}
		return m, nil
	default:
		m.searchInput.Focus()
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		return m, cmd
	}
}

func (m *rootModel) runSearch(q string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		typ := m.app.Config.UI.DefaultType
		res, err := m.app.Search.Search(ctx, q, typ, 0)
		if err != nil {
			return searchErrMsg{err}
		}
		return searchDoneMsg(res)
	}
}

func (m *rootModel) loadStreamsForSelection() tea.Cmd {
	if m.selected == nil {
		return nil
	}
	sel := *m.selected
	m.streamsBusy = true
	m.streamItems = nil
	m.streamsList.SetItems(nil)
	imdb := sel.IMDBID
	metaType := sel.Type
	if metaType == "" {
		metaType = m.app.Config.UI.DefaultType
	}
	season, episode := 0, 0
	if metaType == "series" {
		season, episode = 1, 1
	}
	return func() tea.Msg {
		ctx := context.Background()
		list, err := m.app.Streams.Resolve(ctx, imdb, metaType, season, episode)
		if err != nil {
			return streamsErrMsg{err}
		}
		return streamsDoneMsg(list)
	}
}

func (m *rootModel) updateStreams(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "r":
		if m.selected != nil {
			return m, m.loadStreamsForSelection()
		}
	case "enter":
		if it, ok := m.streamsList.SelectedItem().(streamItem); ok {
			u := it.s.PlayableURL()
			if u != "" {
				_ = clipboard.WriteAll(u)
				m.toast = "Copied to clipboard"
				return m, m.tickToast()
			}
			m.toast = "No URL for this stream"
			return m, m.tickToast()
		}
	}
	var cmd tea.Cmd
	m.streamsList, cmd = m.streamsList.Update(msg)
	return m, cmd
}

func (m *rootModel) updateAddons(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.addonURLMode {
		switch msg.String() {
		case "enter":
			url := strings.TrimSpace(m.addonURL.Value())
			if url == "" {
				return m, nil
			}
			return m, m.installAddon(url)
		default:
			var cmd tea.Cmd
			m.addonURL, cmd = m.addonURL.Update(msg)
			return m, cmd
		}
	}
	switch msg.String() {
	case "a":
		m.addonURLMode = true
		m.addonURL.Focus()
		return m, textinput.Blink
	case "d":
		if it, ok := m.addonList.SelectedItem().(addonItem); ok {
			id := it.a.ID
			if id == "com.linvo.cinemeta" {
				m.toast = "Refusing to remove Cinemeta"
				return m, m.tickToast()
			}
			return m, m.removeAddon(id)
		}
	case "c":
		if it, ok := m.addonList.SelectedItem().(addonItem); ok {
			return m, m.openAddonConfig(it.a)
		}
	}
	var cmd tea.Cmd
	m.addonList, cmd = m.addonList.Update(msg)
	return m, cmd
}

func (m *rootModel) installAddon(url string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		entry, err := m.app.Addons.Install(ctx, url)
		if err != nil {
			return statusMsg{err: err}
		}
		return statusMsg{text: "installed " + entry.ID}
	}
}

func (m *rootModel) removeAddon(id string) tea.Cmd {
	return func() tea.Msg {
		err := m.app.Addons.Remove(id)
		if err != nil {
			return statusMsg{err: err}
		}
		_ = m.app.Reload()
		return statusMsg{text: "removed " + id}
	}
}

func (m *rootModel) openAddonConfig(a config.AddonEntry) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		confURL := strings.TrimSpace(a.ConfigurationURL)
		if confURL == "" {
			man, err := m.app.Client.GetManifest(ctx, a.ManifestURL)
			if err == nil && man.BehaviorHints.OpenURLTemplate != "" {
				confURL = man.BehaviorHints.OpenURLTemplate
			}
		}
		if confURL == "" {
			return statusMsg{err: fmt.Errorf("no configuration URL for this addon")}
		}
		_ = browser.OpenURL(confURL)
		return statusMsg{text: "opened configuration in browser"}
	}
}

func (m *rootModel) updateSettings(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		key := strings.TrimSpace(m.settingsInput.Value())
		if key != "" {
			_ = config.SetKey("tmdb.api_key", key)
			m.settingsInput.SetValue("")
			m.toast = "TMDB key saved"
			_ = m.app.Reload()
			return m, m.tickToast()
		}
	}
	var cmd tea.Cmd
	m.settingsInput, cmd = m.settingsInput.Update(msg)
	return m, cmd
}

func (m *rootModel) View() string {
	if m.width == 0 {
		return "Loading…"
	}
	tabBar := m.renderTabs()
	var body string
	switch m.tab {
	case tabSearch:
		body = lipgloss.JoinVertical(lipgloss.Left,
			m.searchInput.View(),
			"",
			m.searchList.View(),
		)
	case tabStreams:
		head := "Pick a stream, Enter copies URL."
		if m.selected != nil {
			head = fmt.Sprintf("%s (%s) — %s", m.selected.Title, m.selected.IMDBID, m.selected.Type)
		}
		if m.streamsBusy {
			head += "\nLoading…"
		}
		body = lipgloss.JoinVertical(lipgloss.Left, head, "", m.streamsList.View())
	case tabAddons:
		extra := ""
		if m.addonURLMode {
			extra = lipgloss.JoinVertical(lipgloss.Left,
				"",
				"Manifest URL:",
				m.addonURL.View(),
				"(Enter to install, esc cancel)",
			)
		}
		body = lipgloss.JoinVertical(lipgloss.Left,
			"[a] add  [d] remove selected  [c] configure in browser",
			m.addonList.View(),
			extra,
		)
	case tabSettings:
		body = lipgloss.JoinVertical(lipgloss.Left,
			"TMDB API key (optional, improves search):",
			m.settingsInput.View(),
			"",
			"Enter to save. Keys are stored in config.toml.",
		)
	}

	toast := ""
	if m.toast != "" {
		toast = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render(m.toast)
	}
	help := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("tab switch · 1-4 tabs · ctrl+c quit")

	frame := lipgloss.NewStyle().
		MaxWidth(m.width).
		Padding(1, 2).
		Render(lipgloss.JoinVertical(lipgloss.Left,
			tabBar,
			"",
			body,
			"",
			toast,
			help,
		))
	return frame
}

func (m *rootModel) renderTabs() string {
	names := []string{"Search", "Streams", "Addons", "Settings"}
	var parts []string
	for i, n := range names {
		st := lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("252"))
		if i == m.tab {
			st = st.Foreground(lipgloss.Color("205")).Bold(true)
		}
		parts = append(parts, st.Render(fmt.Sprintf("%d:%s", i+1, n)))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

// list items

type titleItem struct{ r search.TitleResult }

func (i titleItem) Title() string       { return i.r.Title }
func (i titleItem) Description() string { return i.r.IMDBID + "  " + i.r.Year + "  " + i.r.Type }
func (i titleItem) FilterValue() string { return i.r.Title }

type streamItem struct{ s streams.ResolvedStream }

func (i streamItem) Title() string {
	t := i.s.Title
	if t == "" {
		t = i.s.Name
	}
	return "[" + i.s.AddonName + "] " + t
}
func (i streamItem) Description() string {
	u := i.s.PlayableURL()
	if len(u) > 120 {
		return u[:117] + "..."
	}
	return u
}
func (i streamItem) FilterValue() string { return i.s.Title + i.s.Name }

type addonItem struct {
	a     config.AddonEntry
	label string
}

func (i addonItem) Title() string       { return i.label }
func (i addonItem) Description() string { return i.a.ManifestURL }
func (i addonItem) FilterValue() string { return i.a.ID }

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
