package main

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
)

type ChannelItem struct {
	channel string
}

func (i ChannelItem) FilterValue() string { return i.channel }
func (i ChannelItem) Title() string {
	if i.channel == "" {
		return "all channels"
	}
	return i.channel
}
func (i ChannelItem) Description() string { return "" }

type VideoItem struct {
	video Video
}

func (i VideoItem) FilterValue() string { return i.video.Title }
func (i VideoItem) Title() string       { return i.video.Title }
func (i VideoItem) Description() string { return "" }

type StatsModel struct {
	videos            []Video
	help              help.Model
	titleSearch       textinput.Model
	channelSelect     list.Model
	videoList         list.Model
	availableChannels []string
	filtered          []Video
	isFiltered        bool
	focusedSearch     int // 0 = none, 1 = title, 2 = channel, 3 = video list
	lastFocused       int // 0 = none, 1 = title, 2 = channel, 3 = video list
	viewMode          int // 0 = rating, 1 = monthly, 2 = video list, 3 = video details
}

type ChannelStats struct {
	Channel    string
	Count      int
	AvgRating  float64
	TotalRated int
}

type MonthStats struct {
	Month string
	Count int
}

type StreakInfo struct {
	VideoCount int
	DaySpan    int
}

type VideoListDelegate struct{}

func (d VideoListDelegate) Height() int                               { return 1 }
func (d VideoListDelegate) Spacing() int                              { return 0 }
func (d VideoListDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d VideoListDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	videoItem, ok := item.(VideoItem)
	if !ok {
		return
	}
	title := truncateString(videoItem.video.Title, 43)
	stars := renderStars(videoItem.video.Rating)
	style := lipgloss.NewStyle().Margin(0).Padding(0, 2)
	if index == m.Index() {
		style = tableSelectedRowStyle.Margin(0).Padding(0, 2)
	}

	// use fixed-width format to ensure perfect alignment
	line := fmt.Sprintf("%-43.43s%6s", title, stars)
	fmt.Fprint(w, style.Render(line))
}

func NewStatsModel() StatsModel {
	titleSearch := textinput.New()
	titleSearch.Placeholder = "search videos..."
	titleSearch.Prompt = "  "
	titleSearch.CharLimit = 50
	titleSearch.Width = 50

	channelSelect := list.New([]list.Item{}, list.NewDefaultDelegate(), 50, 1)
	channelSelect.SetShowStatusBar(false)
	channelSelect.SetFilteringEnabled(false)
	channelSelect.SetShowTitle(false)
	channelSelect.SetShowHelp(false)

	// Create video list with custom delegate
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = tableSelectedRowStyle.Bold(false).Margin(0, 0).Padding(0, 0)
	delegate.Styles.SelectedDesc = tableSelectedRowStyle.Bold(false).Margin(0, 0).Padding(0, 0)
	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.Margin(0, 0).Padding(0, 0)
	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.Margin(0, 0).Padding(0, 0)

	videoList := list.New([]list.Item{}, VideoListDelegate{}, 50, 12)
	videoList.SetShowStatusBar(false)
	videoList.SetFilteringEnabled(false)
	videoList.SetShowTitle(false)
	videoList.SetShowHelp(false)
	videoList.SetShowPagination(false)

	h := help.New()
	h.ShowAll = false

	return StatsModel{
		help:          h,
		titleSearch:   titleSearch,
		channelSelect: channelSelect,
		videoList:     videoList,
		filtered:      []Video{},
		isFiltered:    false,
		focusedSearch: 0,
		lastFocused:   0,
		viewMode:      0,
	}
}

func (m StatsModel) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		func() tea.Msg {
			videos, err := loadVideos()
			if err != nil {
				return err
			}
			return LoadVideosMsg{videos: videos}
		},
	)
}

func (m *StatsModel) updateChannelList() {
	// get unique channels with their counts
	channelMap := make(map[string]int)
	for _, video := range m.videos {
		channel := getVideoChannel(video)
		channelMap[channel]++
	}

	// convert to slice and sort by count (most logged first)
	type channelCount struct {
		name  string
		count int
	}

	var channelCounts []channelCount
	for channel, count := range channelMap {
		channelCounts = append(channelCounts, channelCount{name: channel, count: count})
	}

	// sort by count, then by name for ties
	sort.Slice(channelCounts, func(i, j int) bool {
		if channelCounts[i].count == channelCounts[j].count {
			return channelCounts[i].name < channelCounts[j].name
		}
		return channelCounts[i].count > channelCounts[j].count
	})

	channels := []string{""}
	for _, cc := range channelCounts {
		channels = append(channels, cc.name)
	}

	// update list items
	items := make([]list.Item, len(channels))
	for i, channel := range channels {
		items[i] = ChannelItem{channel: channel}
	}

	m.channelSelect.SetItems(items)
	m.availableChannels = channels
}

func (m *StatsModel) getSelectedChannel() string {
	if selectedItem := m.channelSelect.SelectedItem(); selectedItem != nil {
		if channelItem, ok := selectedItem.(ChannelItem); ok {
			return channelItem.channel
		}
	}
	return ""
}

func (m *StatsModel) filterStats() {
	titleQuery := strings.TrimSpace(m.titleSearch.Value())
	selectedChannel := m.getSelectedChannel()

	if titleQuery == "" && selectedChannel == "" {
		m.isFiltered = false
		m.filtered = m.videos
		m.updateVideoList()
		return
	}

	m.isFiltered = true
	m.filtered = []Video{}

	for _, video := range m.videos {
		matchesTitle := true
		matchesChannel := true

		// fuzzy find title
		if titleQuery != "" {
			searchable := []string{strings.ToLower(video.Title)}
			matches := fuzzy.Find(titleQuery, searchable)
			matchesTitle = len(matches) > 0
		}

		// make sure log is from selected channel
		if selectedChannel != "" {
			videoChannel := getVideoChannel(video)
			matchesChannel = videoChannel == selectedChannel
		}

		if matchesTitle && matchesChannel {
			m.filtered = append(m.filtered, video)
		}
	}
	m.updateVideoList()
}

func (m *StatsModel) calculateStats() (
	totalVideos int, avgRating float64, totalRated int,
	rewatchCount int, channelStats []ChannelStats,
	monthStats []MonthStats, ratingDist map[float64]int,
) {
	videosToUse := m.videos
	if m.isFiltered {
		videosToUse = m.filtered
	}

	totalVideos = len(videosToUse)
	if totalVideos == 0 {
		return
	}

	var totalRatingSum float64
	channelMap := make(map[string]*ChannelStats)
	monthMap := make(map[string]int)
	ratingDist = make(map[float64]int)

	// init rating distribution
	for i := 1.0; i <= 5.0; i += 0.5 {
		ratingDist[i] = 0
	}

	for _, video := range videosToUse {
		if video.Rating > 0 {
			totalRatingSum += video.Rating
			totalRated++
			ratingDist[video.Rating]++
		}

		if video.Rewatched {
			rewatchCount++
		}

		// channel stats
		channel := getVideoChannel(video)
		if _, exists := channelMap[channel]; !exists {
			channelMap[channel] = &ChannelStats{Channel: channel}
		}
		stats := channelMap[channel]
		stats.Count++
		if video.Rating > 0 {
			stats.TotalRated++
			currentSum := stats.AvgRating * float64(stats.TotalRated-1)
			stats.AvgRating = (currentSum + video.Rating) / float64(stats.TotalRated)
		}

		// month stats
		if video.LogDate != "" {
			if logTime, err := time.Parse(DateTimeFormat, video.LogDate); err == nil {
				monthKey := logTime.Format(MonthFormat)
				monthMap[monthKey]++
			}
		}
	}

	if totalRated > 0 {
		avgRating = totalRatingSum / float64(totalRated)
	}

	// convert and sort channel stats
	for _, stats := range channelMap {
		channelStats = append(channelStats, *stats)
	}
	// most logged first
	sort.Slice(channelStats, func(i, j int) bool {
		return channelStats[i].Count > channelStats[j].Count
	})

	// convert and sort month stats
	for month, count := range monthMap {
		monthStats = append(monthStats, MonthStats{Month: month, Count: count})
	}
	// most recent first
	sort.Slice(monthStats, func(i, j int) bool {
		return monthStats[i].Month > monthStats[j].Month
	})

	return
}

func (m *StatsModel) renderStars(rating float64) string {
	ratingStr := ""
	if rating > 0 {
		for j := 1; j <= 5; j++ {
			starValue := float64(j)
			if rating >= starValue {
				ratingStr += "★" // filled star
			} else if rating >= starValue-0.5 {
				ratingStr += "⯨" // half star
			}
		}
	}

	return ratingStr
}

func (m *StatsModel) getStreaks() (StreakInfo, StreakInfo) {
	videosToUse := m.videos
	if m.isFiltered {
		videosToUse = m.filtered
	}

	if len(videosToUse) == 0 {
		return StreakInfo{0, 0}, StreakInfo{0, 0}
	}

	var dateGroups []struct {
		date  time.Time
		count int
	}

	for _, video := range videosToUse {
		if video.LogDate == "" {
			continue
		}

		logTime, err := time.Parse(DateTimeFormat, video.LogDate)
		if err != nil {
			continue
		}

		// truncate to date for comparison
		logDate := logTime.Truncate(24 * time.Hour)

		// date exists already
		if len(dateGroups) > 0 && dateGroups[len(dateGroups)-1].date.Equal(logDate) {
			dateGroups[len(dateGroups)-1].count++
		} else {
			dateGroups = append(dateGroups, struct {
				date  time.Time
				count int
			}{logDate, 1})
		}
	}

	if len(dateGroups) == 0 {
		return StreakInfo{0, 0}, StreakInfo{0, 0}
	}

	// calculate current streak
	var currentStreak StreakInfo
	today := time.Now().Truncate(24 * time.Hour)
	mostRecentDate := dateGroups[0].date

	// check if recent date is today or yesterday
	daysSinceLastWatch := int(today.Sub(mostRecentDate).Hours() / 24)
	if daysSinceLastWatch <= 1 {
		currentStreak.VideoCount = dateGroups[0].count
		currentStreak.DaySpan = 1

		for i := 1; i < len(dateGroups); i++ {
			prevDate := dateGroups[i-1].date
			currDate := dateGroups[i].date

			// check if dates are consecutive
			daysDiff := int(prevDate.Sub(currDate).Hours() / 24)
			if daysDiff == 1 {
				currentStreak.VideoCount += dateGroups[i].count
				currentStreak.DaySpan++
			} else {
				break
			}
		}
	}

	// calculate best streak
	var longestStreak StreakInfo
	var tempStreak StreakInfo

	for i := range dateGroups {
		if i == 0 {
			tempStreak = StreakInfo{VideoCount: dateGroups[i].count, DaySpan: 1}
		} else {
			prevDate := dateGroups[i-1].date
			currDate := dateGroups[i].date

			// check if dates are consecutive
			daysDiff := int(prevDate.Sub(currDate).Hours() / 24)
			if daysDiff == 1 {
				tempStreak.VideoCount += dateGroups[i].count
				tempStreak.DaySpan++
			} else {
				// streak broken
				if tempStreak.VideoCount > longestStreak.VideoCount {
					longestStreak = tempStreak
				}
				tempStreak = StreakInfo{VideoCount: dateGroups[i].count, DaySpan: 1}
			}
		}
	}
	if tempStreak.VideoCount > longestStreak.VideoCount {
		longestStreak = tempStreak
	}

	return currentStreak, longestStreak
}

func (m *StatsModel) getDasboardStrings(totalVideos int, avgRating float64,
	totalRated int, rewatchCount int, channelStats []ChannelStats,
) (string, string, string, string) {
	totalCard := fmt.Sprintf(" Videos\n%d total", totalVideos)
	avgCard := ""
	if totalRated > 0 {
		avgCard = fmt.Sprintf(" Rating\n%.1f/5", avgRating)
	} else {
		avgCard = " Rating\n"
	}
	rewatchCard := fmt.Sprintf(" Rewatch\n%.0f%%", float64(rewatchCount)/float64(totalVideos)*100)
	channelCountCard := ""
	if m.getSelectedChannel() != "" {
		selectedChannel := m.getSelectedChannel()
		// find the selected channel's stats
		var channelInfo string
		for _, stats := range channelStats {
			if stats.Channel == selectedChannel {
				if stats.TotalRated > 0 {
					channelInfo = fmt.Sprintf(" %d (%.1f )", stats.Count, stats.AvgRating)
				} else {
					channelInfo = fmt.Sprintf(" %d", stats.Count)
				}
				break
			}
		}
		channelCountCard = fmt.Sprintf(" Channel\n%s", channelInfo)
	} else {
		channelCountCard = fmt.Sprintf(" Channels\n%d unique", len(channelStats))
	}

	return totalCard, avgCard, rewatchCard, channelCountCard
}

func (m *StatsModel) renderDashboardCards(str1 string, str2 string, str3 *string, str4 *string) string {
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Height(2)

	if str3 == nil || str4 == nil {
		card1 := cardStyle.Width(26).Render(str1)
		card2 := cardStyle.Width(28).Render(str2)
		return lipgloss.JoinHorizontal(lipgloss.Top, card1, card2)
	}

	card1 := cardStyle.Width(12).Render(str1)
	card2 := cardStyle.Width(12).Render(str2)
	card3 := cardStyle.Width(12).Render(*str3)
	card4 := cardStyle.Width(14).Render(*str4)

	return lipgloss.JoinHorizontal(lipgloss.Top, card1, card2, card3, card4)
}

func (m *StatsModel) updateVideoList() {
	videosToUse := m.videos
	if m.isFiltered {
		videosToUse = m.filtered
	}

	items := make([]list.Item, len(videosToUse))
	for i, video := range videosToUse {
		items[i] = VideoItem{video: video}
	}

	m.videoList.SetItems(items)
}

func (m *StatsModel) setFocus(target int) {
	// Always blur current inputs first
	m.titleSearch.Blur()

	// Update state
	if target != 0 {
		m.lastFocused = target
	}
	m.focusedSearch = target

	// Focus the appropriate input
	if target == 1 {
		m.titleSearch.Focus()
	}
}

func (m *StatsModel) toggleSearch() {
	if m.focusedSearch == 0 {
		// Return to last focused search, default to title search
		target := m.lastFocused
		if target == 0 {
			target = 1
		}
		m.setFocus(target)
	} else {
		m.setFocus(0)
	}
}

func (m StatsModel) cycleField(field *int, forward bool, numModes int) (next int) {
	if forward {
		next = (*field + 1) % numModes
	} else {
		next = (*field + numModes - 1) % numModes
	}

	return next
}

func (m StatsModel) Update(msg tea.Msg) (StatsModel, tea.Cmd) {
	var titleCmd, channelCmd, listCmd tea.Cmd

	switch msg := msg.(type) {
	case LoadVideosMsg:
		m.videos = msg.videos
		m.filtered = msg.videos
		m.updateChannelList()
		m.updateVideoList()
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, GlobalKeyMap.Help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, GlobalKeyMap.Search):
			m.toggleSearch()
		case key.Matches(msg, GlobalKeyMap.Cycle):
			next := m.cycleField(&m.focusedSearch, true, 3)
			m.setFocus(next)
		case key.Matches(msg, GlobalKeyMap.CycleBack):
			next := m.cycleField(&m.focusedSearch, false, 3)
			m.setFocus(next)
		case key.Matches(msg, GlobalKeyMap.SearchBack):
			if m.focusedSearch > 0 {
				m.setFocus(0)
				m.filterStats()
			}
		case m.focusedSearch == 1: // title search
			m.titleSearch, titleCmd = m.titleSearch.Update(msg)
			m.filterStats()
			m.updateVideoList()
			return m, titleCmd
		case m.focusedSearch == 2: // channel select
			// ignore left right
			if key.Matches(msg, GlobalKeyMap.Left) || key.Matches(msg, GlobalKeyMap.Right) {
				return m, nil
			}
			m.channelSelect, channelCmd = m.channelSelect.Update(msg)
			m.filterStats()
			m.updateVideoList()
			return m, channelCmd
		case m.focusedSearch == 0: // chart view
			switch {
			case key.Matches(msg, GlobalKeyMap.Left): // switch between chart views
				m.viewMode = m.cycleField(&m.viewMode, false, 3)
			case key.Matches(msg, GlobalKeyMap.Right):
				m.viewMode = m.cycleField(&m.viewMode, true, 3)
			case m.viewMode == 2: // video list
				switch {
				case key.Matches(msg, GlobalKeyMap.Up), key.Matches(msg, GlobalKeyMap.Down):
					m.videoList, listCmd = m.videoList.Update(msg)
					return m, listCmd
				case key.Matches(msg, GlobalKeyMap.Select):
					return m.handleVideoSelection()
				case key.Matches(msg, GlobalKeyMap.Edit):
					videosToUse := m.videos
					if m.isFiltered {
						videosToUse = m.filtered
					}
					if len(videosToUse) > 0 {
						selectedItem := m.videoList.SelectedItem()
						if videoItem, ok := selectedItem.(VideoItem); ok {
							return m, func() tea.Msg {
								return NavigateMsg{View: LogVideoView, VideoID: videoItem.video.ID}
							}
						}
					}
				}
			default:
				m.videoList, listCmd = m.videoList.Update(msg)
				return m, listCmd
			}
		}
	}

	return m, nil
}

func (m StatsModel) View() string {
	var s strings.Builder

	s.WriteString(headerStyle.Render("video stats") + "\n")

	searchBoxStyle := searchStyle.Width(26)
	channelSelectStyle := searchStyle.Width(28)
	// apply focus styling
	if m.focusedSearch == 1 {
		searchBoxStyle = searchBoxStyle.BorderForeground(primaryColor)
	}
	if m.focusedSearch == 2 {
		channelSelectStyle = channelSelectStyle.BorderForeground(primaryColor)
	}

	// search box content
	searchBox := searchBoxStyle.Render(m.titleSearch.View())
	// channel select content
	channelSelectContent := " all channels"
	if selectedItem := m.channelSelect.SelectedItem(); selectedItem != nil {
		if channelItem, ok := selectedItem.(ChannelItem); ok {
			channelSelectContent = " " + channelItem.Title()
		}
	}
	channelSelectBox := channelSelectStyle.Render(channelSelectContent)

	// combine together
	searchRow := lipgloss.JoinHorizontal(lipgloss.Top, searchBox, channelSelectBox)
	s.WriteString("\n" + searchRow + "\n")

	if m.isFiltered {
		percentage := float64(len(m.filtered)) / float64(len(m.videos)) * 100
		filterInfo := fmt.Sprintf("Filtered: %d/%d (%.0f%%)", len(m.filtered), len(m.videos), percentage)
		s.WriteString(descriptionStyle.Render(filterInfo) + "\n")
	} else {
		s.WriteString(descriptionStyle.Render("Filtered: All videos") + "\n")
	}

	// NOTE :: calculate stats
	totalVideos, avgRating, totalRated, rewatchCount, channelStats, monthStats, ratingDist := m.calculateStats()

	if totalVideos == 0 {
		s.WriteString(centerHorizontally("\n no videos logged yet \n", 60))
		s.WriteString("\n" + m.help.View(StatsKeyMap{}))
		return s.String()
	}

	// streak cards
	currentStreak, longestStreak := m.getStreaks()
	currenStreakCard := fmt.Sprintf("Current streak: \n%d videos in %d days", currentStreak.VideoCount, currentStreak.DaySpan)
	longestStreakCard := fmt.Sprintf("Best streak: \n%d videos in %d days", longestStreak.VideoCount, longestStreak.DaySpan)
	streakRow := m.renderDashboardCards(longestStreakCard, currenStreakCard, nil, nil)
	s.WriteString("\n" + streakRow + "\n")

	// dashboard cards
	totalCard, avgCard, rewatchCard, channelCountCard := m.getDasboardStrings(totalVideos, avgRating, totalRated, rewatchCount, channelStats)
	row := m.renderDashboardCards(totalCard, avgCard, &rewatchCard, &channelCountCard)
	s.WriteString("\n" + row + "\n")

	// show selected chart
	if m.viewMode == 0 {
		s.WriteString(m.renderChart(m.prepareRatingChartData(ratingDist), m.focusedSearch == 0))
	} else if m.viewMode == 1 {
		s.WriteString(m.renderChart(m.prepareMonthlyChartData(monthStats), m.focusedSearch == 0))
	} else {
		s.WriteString(m.renderVideoList())
	}

	// show compact channels if `all channels`
	if len(channelStats) > 0 && m.getSelectedChannel() == "" {
		s.WriteString(m.renderCompactChannels(channelStats))
	} else {
		if !m.help.ShowAll {
			s.WriteString("\n\n\n\n\n") // more padding when compact
		} else {
			s.WriteString("\n") // maintain consistent height
		}
	}

	// help section
	keymap := StatsKeyMap{}
	s.WriteString("\n" + m.help.View(keymap))

	return s.String()
}

func (m StatsModel) renderVideoList() string {
	listStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Margin(1, 0).
		Width(56)

	if m.focusedSearch == 0 && m.viewMode == 2 {
		listStyle = listStyle.BorderForeground(primaryColor)
	}

	var content strings.Builder
	videosToUse := m.videos
	if m.isFiltered {
		videosToUse = m.filtered
	}
	content.WriteString(fmt.Sprintf("   Videos (%d/%d)\n", m.videoList.Index()+1, len(videosToUse)))

	if len(m.filtered) == 0 {
		content.WriteString("No videos to display")
	} else {
		content.WriteString(m.videoList.View())
	}

	return listStyle.Render(content.String()) + "\n"
}

func (m StatsModel) handleVideoSelection() (StatsModel, tea.Cmd) {
	videosToUse := m.videos
	if m.isFiltered {
		videosToUse = m.filtered
	}

	if len(videosToUse) == 0 {
		return m, nil
	}

	selectedItem := m.videoList.SelectedItem()
	if videoItem, ok := selectedItem.(VideoItem); ok {
		return m, func() tea.Msg {
			return NavigateMsg{View: LogDetailsView, VideoID: videoItem.video.ID}
		}
	}
	return m, nil
}

func (m StatsModel) renderCompactChannels(channelStats []ChannelStats) string {
	listStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Width(56)

	var list strings.Builder

	list.WriteString(" Top Channels: ")

	limit := min(len(channelStats), 3)

	for i, stats := range channelStats[:limit] {
		if i > 0 {
			list.WriteString("\n\t\t\t\t• ")
		} else {
			list.WriteString("• ")
		}
		avgStr := ""
		if stats.TotalRated > 0 {
			avgStr = fmt.Sprintf("(%.1f)", stats.AvgRating)
		}
		list.WriteString(fmt.Sprintf("%-15s  %-3d%-6s", stats.Channel, stats.Count, avgStr))
	}

	return listStyle.Render(list.String()) + "\n"
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "…"
}

type StatsKeyMap struct{}

func (k StatsKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		GlobalKeyMap.Cycle,
		GlobalKeyMap.Search,
		GlobalKeyMap.Left,
		GlobalKeyMap.Right,
		GlobalKeyMap.Help,
	}
}

func (k StatsKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			GlobalKeyMap.Cycle,
			GlobalKeyMap.CycleBack,
			GlobalKeyMap.Search,
		},
		{
			GlobalKeyMap.Left,
			GlobalKeyMap.Right,
			GlobalKeyMap.Up,
			GlobalKeyMap.Down,
		},
		{
			GlobalKeyMap.Help,
			GlobalKeyMap.Select,
			GlobalKeyMap.Edit,
			GlobalKeyMap.Back,
			GlobalKeyMap.Exit,
		},
	}
}

func getVideoChannel(video Video) string {
	if video.Channel == "" {
		return "Unknown Channel"
	}
	return video.Channel
}
