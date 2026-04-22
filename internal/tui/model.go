package tui

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"bm/internal/app"
	"bm/internal/search"
	"bm/internal/streams"
	"bm/internal/stremio"
)

// rootModel is the Bubble Tea root; use pointer receiver so updates persist.
type rootModel struct {
	app *app.App

	width, height int
	tab           int

	searchInput textinput.Model
	searchList  list.Model
	searchItems []list.Item

	searchActive bool
	browseMode   int // browsePopular | browseFeatured | browseSearch

	streamsList  list.Model
	episodesList list.Model
	selected     *search.TitleResult
	streamsBusy  bool
	episodesBusy bool
	streamsStage int // stageEpisodes | stageStreams
	seasonPick   int
	episodePick  int
	seriesMeta   *stremio.Meta

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

	toast    string
	helpOpen bool
}

func newRootModel(ap *app.App) *rootModel {
	si := textinput.New()
	si.Placeholder = "Search movies & series…"
	si.CharLimit = 200
	si.Width = 50
	si.Blur()

	sl := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	sl.Title = "Popular (Cinemeta)"
	sl.SetShowStatusBar(false)
	sl.SetFilteringEnabled(false)
	sl.DisableQuitKeybindings()

	strl := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	strl.Title = "Streams"
	strl.SetShowStatusBar(false)
	strl.SetFilteringEnabled(false)
	strl.DisableQuitKeybindings()

	epL := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	epL.Title = "Episodes"
	epL.SetShowStatusBar(false)
	epL.DisableQuitKeybindings()
	epL.SetFilteringEnabled(false)

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
		episodesList:    epL,
		addonList:       al,
		addonURL:        aui,
		settingsInput:   tmdb,
		searchMediaType: normalizeSearchMediaType(ap.Config.UI.DefaultType),
		streamListOrder: streams.NormalizeOrder(ap.Config.UI.StreamOrder),
		browseMode:      browsePopular,
		streamsStage:    stageEpisodes,
	}
	m.refreshAddonList()
	return m
}

func (m *rootModel) Init() tea.Cmd {
	return m.runCinemetaPopular()
}

func (m *rootModel) tickToast() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return toastClearMsg{}
	})
}

// globalKeysBlocked is true when top-level shortcuts (1–4, tab) would steal keys from a focused text field.
func (m *rootModel) globalKeysBlocked() bool {
	if m.tab == tabSearch && m.searchActive && m.searchInput.Focused() {
		return true
	}
	if m.tab == tabAddons && m.addonURLMode && m.addonURL.Focused() {
		return true
	}
	return false
}

func (m *rootModel) effectiveMetaType() string {
	if m.selected != nil {
		t := strings.TrimSpace(m.selected.Type)
		if t == "series" || t == "movie" {
			return t
		}
	}
	return normalizeSearchMediaType(m.app.Config.UI.DefaultType)
}

func (m *rootModel) closeSearchInput() {
	m.searchActive = false
	m.searchInput.Blur()
	m.searchInput.SetValue("")
}

func (m *rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		h := max(6, m.height-8)
		m.searchList.SetSize(max(20, m.width-4), h)
		m.streamsList.SetSize(max(20, m.width-4), h)
		m.episodesList.SetSize(max(20, m.width-4), h)
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

		if ks == "?" || msg.Type == tea.KeyF1 {
			m.helpOpen = !m.helpOpen
			return m, nil
		}

		if ks == "q" && !m.globalKeysBlocked() {
			return m, tea.Quit
		}

		if ks == "esc" {
			if m.addonURLMode {
				m.addonURLMode = false
				m.addonURL.Blur()
				return m, nil
			}
			if m.tab == tabSearch && m.searchActive {
				m.closeSearchInput()
				return m, nil
			}
			if m.tab == tabStreams {
				return m.backFromStreams()
			}
			return m, nil
		}

		if ks == "ctrl+c" {
			return m, tea.Quit
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
		m.searchItems = make([]list.Item, 0, len(msg.items))
		for _, r := range msg.items {
			m.searchItems = append(m.searchItems, titleItem{r: r})
		}
		m.searchList.SetItems(m.searchItems)
		m.browseMode = msg.source
		switch msg.source {
		case browsePopular:
			m.searchList.Title = "Popular (Cinemeta)"
		case browseFeatured:
			m.searchList.Title = "Featured (Cinemeta)"
		default:
			m.searchList.Title = "Search results"
		}
		m.toast = fmt.Sprintf("%d results", len(msg.items))
		return m, m.tickToast()

	case searchErrMsg:
		m.toast = "browse: " + msg.err.Error()
		return m, m.tickToast()

	case metaDoneMsg:
		m.episodesBusy = false
		if m.selected == nil {
			return m, nil
		}
		m.seriesMeta = msg.meta
		items := buildEpisodeListItems(msg.meta.Videos)
		m.episodesList.SetItems(items)
		for i, it := range items {
			if _, ok := it.(episodeItem); ok {
				m.episodesList.Select(i)
				break
			}
		}
		if len(items) == 0 {
			m.toast = "no episodes in catalog meta"
			return m, m.tickToast()
		}
		m.toast = fmt.Sprintf("%d rows", len(items))
		return m, m.tickToast()

	case metaErrMsg:
		m.episodesBusy = false
		errStr := msg.err.Error()
		_, cmd := m.backToBrowse()
		m.toast = "meta: " + errStr
		return m, tea.Batch(cmd, m.tickToast())

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
