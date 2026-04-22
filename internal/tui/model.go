package tui

import (
	"fmt"
	"slices"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"bm/internal/app"
	"bm/internal/search"
	"bm/internal/streams"
)

// rootModel is the Bubble Tea root; use pointer receiver so updates persist.
type rootModel struct {
	app *app.App

	width, height int
	tab           int

	searchInput textinput.Model
	searchList  list.Model
	searchItems []list.Item

	streamsList list.Model
	selected    *search.TitleResult
	streamsBusy bool

	allResolvedStreams []streams.ResolvedStream
	streamAddonTabs    []streamAddonTab
	streamsAddonTabIdx int
	streamListOrder    string

	addonList    list.Model
	addonItems   []list.Item
	addonURLMode bool
	addonURL     textinput.Model

	settingsInput textinput.Model

	searchMediaType string
	lastSearchQuery string

	toast       string
	helpOpen    bool
	quitConfirm bool
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
		app:             ap,
		searchInput:     si,
		searchList:      sl,
		streamsList:     strl,
		addonList:       al,
		addonURL:        aui,
		settingsInput:   tmdb,
		searchMediaType: normalizeSearchMediaType(ap.Config.UI.DefaultType),
		streamListOrder: streams.NormalizeOrder(ap.Config.UI.StreamOrder),
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

// globalKeysBlocked is true when top-level shortcuts (1–4, tab) would steal keys from a focused text field.
func (m *rootModel) globalKeysBlocked() bool {
	if m.tab == tabSearch && m.searchInput.Focused() {
		return true
	}
	if m.tab == tabAddons && m.addonURLMode && m.addonURL.Focused() {
		return true
	}
	return false
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
		ks := msg.String()

		if m.helpOpen {
			switch ks {
			case "esc", "?", "q":
				m.helpOpen = false
				return m, nil
			case "ctrl+c":
				m.helpOpen = false
				return m, nil
			}
			return m, nil
		}

		if m.quitConfirm {
			switch ks {
			case "y", "Y":
				return m, tea.Quit
			case "n", "N", "esc":
				m.quitConfirm = false
				return m, nil
			case "ctrl+c":
				return m, tea.Quit
			}
			return m, nil
		}

		if ks == "?" || msg.Type == tea.KeyF1 {
			m.quitConfirm = false
			m.helpOpen = !m.helpOpen
			return m, nil
		}

		if ks == "esc" {
			if m.addonURLMode {
				m.addonURLMode = false
				m.addonURL.Blur()
				return m, nil
			}
			if m.tab == tabStreams {
				return m.backFromStreams()
			}
			m.quitConfirm = true
			return m, nil
		}

		if ks == "ctrl+c" {
			m.quitConfirm = true
			return m, nil
		}

		if !m.globalKeysBlocked() {
			switch ks {
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
		if m.selected == nil {
			return m, nil
		}
		m.allResolvedStreams = slices.Clone(msg)
		streams.ApplySort(m.allResolvedStreams, m.streamListOrder)
		m.streamAddonTabs = buildStreamAddonTabs(m.allResolvedStreams)
		m.streamsAddonTabIdx = 0
		n := m.applyStreamsAddonFilter()
		m.toast = fmt.Sprintf("%d streams", n)
		return m, m.tickToast()

	case streamsErrMsg:
		m.streamsBusy = false
		if m.selected == nil {
			return m, nil
		}
		m.allResolvedStreams = nil
		m.streamAddonTabs = nil
		m.streamsAddonTabIdx = 0
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
		m.streamListOrder = streams.NormalizeOrder(m.app.Config.UI.StreamOrder)
		if len(m.allResolvedStreams) > 0 {
			streams.ApplySort(m.allResolvedStreams, m.streamListOrder)
			_ = m.applyStreamsAddonFilter()
		}
		m.refreshAddonList()
		return m, m.tickToast()

	case toastClearMsg:
		m.toast = ""
		return m, nil
	}

	return m, nil
}
