package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

type ChartData struct {
	Title    string
	Labels   []string
	Values   []int
	MaxItems int
}

func (m StatsModel) renderChart(data ChartData, isFocused bool) string {
	chartStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Margin(1, 0).
		Width(56)

	if isFocused {
		chartStyle = chartStyle.BorderForeground(primaryColor)
	}

	var chart strings.Builder
	chart.WriteString(" " + data.Title + "\n\n")

	if len(data.Values) == 0 {
		chart.WriteString("No data available")
		return chartStyle.Render(chart.String()) + "\n"
	}

	// Find max count for scaling
	maxCount := 0
	for _, count := range data.Values {
		if count > maxCount {
			maxCount = count
		}
	}

	maxBarHeight := 8

	// Limit items if needed
	items := len(data.Values)
	if data.MaxItems > 0 && items > data.MaxItems {
		items = data.MaxItems
	}

	// Build each row of the chart from top to bottom
	for row := maxBarHeight; row >= 1; row-- {
		for i := 0; i < items; i++ {
			count := data.Values[i]
			barHeight := 0
			if maxCount > 0 && count > 0 {
				barHeight = int(float64(maxBarHeight) * float64(count) / float64(maxCount))
				if barHeight == 0 && count > 0 {
					barHeight = 1 // Ensure at least 1 row for non-zero counts
				}
			}

			if row <= barHeight {
				chart.WriteString("█████ ")
			} else {
				chart.WriteString("      ")
			}
		}
		chart.WriteString("\n")
	}

	// Add labels
	for i := 0; i < items; i++ {
		chart.WriteString(fmt.Sprintf("%-6s", data.Labels[i]))
	}
	chart.WriteString("\n")

	// Add values
	for i := 0; i < items; i++ {
		chart.WriteString(fmt.Sprintf("  %-4s", fmt.Sprintf("%d", data.Values[i])))
	}
	chart.WriteString("\n")

	return chartStyle.Render(chart.String()) + "\n"
}

func (m StatsModel) prepareRatingChartData(ratingDist map[float64]int) ChartData {
	ratings := []float64{1.0, 1.5, 2.0, 2.5, 3.0, 3.5, 4.0, 4.5, 5.0}

	labels := make([]string, len(ratings))
	values := make([]int, len(ratings))

	for i, rating := range ratings {
		labels[i] = m.renderStars(rating)
		values[i] = ratingDist[rating]
	}

	return ChartData{
		Title:    "  Ratings",
		Labels:   labels,
		Values:   values,
		MaxItems: 0, // Show all ratings
	}
}

func (m StatsModel) prepareMonthlyChartData(monthStats []MonthStats) ChartData {
	if len(monthStats) == 0 {
		return ChartData{
			Title:    "  Months",
			Labels:   []string{},
			Values:   []int{},
			MaxItems: 9,
		}
	}

	// Create a map for quick lookup of existing month data
	monthMap := make(map[string]int)
	for _, month := range monthStats {
		monthMap[month.Month] = month.Count
	}

	// Generate continuous months from oldest to newest (up to 9 months)
	var continuousMonths []string

	// Start from the most recent month and go back
	now := time.Now()
	for i := 0; i < 9; i++ {
		monthTime := now.AddDate(0, -i, 0)
		monthKey := monthTime.Format(MonthFormat)
		continuousMonths = append([]string{monthKey}, continuousMonths...) // Prepend to reverse order
	}

	// Create labels and values arrays
	labels := make([]string, len(continuousMonths))
	values := make([]int, len(continuousMonths))

	for i, monthKey := range continuousMonths {
		// Set the label to the month key (MM/YY format)
		labels[i] = monthKey

		// Get count from map, or 0 if month doesn't exist
		if count, exists := monthMap[monthKey]; exists {
			values[i] = count
		} else {
			values[i] = 0
		}
	}

	return ChartData{
		Title:    "  Months",
		Labels:   labels,
		Values:   values,
		MaxItems: 9,
	}
}
