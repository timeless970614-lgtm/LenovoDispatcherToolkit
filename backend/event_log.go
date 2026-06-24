//go:build windows

package backend

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// EventLogSummary holds the summary of System event log capture
type EventLogSummary struct {
	TotalEvents   int              `json:"totalEvents"`
	TimeRange     string           `json:"timeRange"`
	CriticalCount int              `json:"criticalCount"`
	ErrorCount    int              `json:"errorCount"`
	WarningCount  int              `json:"warningCount"`
	InfoCount     int              `json:"infoCount"`
	TopProviders  []EventProvider  `json:"topProviders"`
	RecentErrors  []EventLogEntry  `json:"recentErrors"`
	RecentEvents  []EventLogEntry  `json:"recentEvents"`
	RawOutput     string           `json:"rawOutput"`
}

// EventProvider holds provider stats
type EventProvider struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// EventLogEntry holds a single event
type EventLogEntry struct {
	Time          string `json:"time"`
	Level         string `json:"level"`
	ProviderName  string `json:"providerName"`
	EventID       int    `json:"eventId"`
	Message       string `json:"message"`
}

// CaptureSystemEventLog captures System event log entries
// hoursBack: how many hours to look back (0 = use maxEvents limit)
// maxEvents: max events to return (0 = default 500)
func CaptureSystemEventLog(hoursBack int, maxEvents int) EventLogSummary {
	if maxEvents <= 0 {
		maxEvents = 500
	}

	summary := EventLogSummary{
		TimeRange: fmt.Sprintf("Last %d hours", hoursBack),
	}

	// Build PowerShell command to get System event log
	var psScript string
	if hoursBack > 0 {
		psScript = fmt.Sprintf(
			`$start = (Get-Date).AddHours(-%d); `+
				`$events = Get-WinEvent -LogName System -MaxEvents %d | Where-Object { $_.TimeCreated -ge $start }; `+
				`$events | Select-Object TimeCreated, LevelDisplayName, ProviderName, Id, Message | ConvertTo-Json -Depth 2`,
			hoursBack, maxEvents*10)
		summary.TimeRange = fmt.Sprintf("Last %d hours (up to %d events)", hoursBack, maxEvents)
	} else {
		psScript = fmt.Sprintf(
			`Get-WinEvent -LogName System -MaxEvents %d | `+
				`Select-Object TimeCreated, LevelDisplayName, ProviderName, Id, Message | ConvertTo-Json -Depth 2`,
			maxEvents)
		summary.TimeRange = fmt.Sprintf("Latest %d events", maxEvents)
	}

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Fall back to wevtutil if PowerShell fails
		summary.RawOutput = fmt.Sprintf("PowerShell failed: %v\nOutput: %s", err, string(output))
		return tryWevtUtilFallback(hoursBack, maxEvents, summary)
	}

	rawJSON := strings.TrimSpace(string(output))
	summary.RawOutput = rawJSON

	// Parse JSON array of events
	parseEventLogJSON(rawJSON, &summary)

	// Try wevtutil as supplementary (more reliable for count)
	enrichWithWevtUtil(&summary)

	return summary
}

// parseEventLogJSON parses JSON output from Get-WinEvent
func parseEventLogJSON(raw string, summary *EventLogSummary) {
	// Handle single event case (not wrapped in array)
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, "[") {
		raw = "[" + raw + "]"
	}

	var events []map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &events); err != nil {
		// Try parsing line by line if array parse fails
		return
	}

	providerCount := make(map[string]int)

	for _, ev := range events {
		summary.TotalEvents++

		level := getStringField(ev, "LevelDisplayName")
		switch level {
		case "Critical", "严重":
			summary.CriticalCount++
		case "Error", "错误":
			summary.ErrorCount++
		case "Warning", "警告":
			summary.WarningCount++
		default:
			summary.InfoCount++
		}

		provider := getStringField(ev, "ProviderName")
		providerCount[provider]++

		entry := EventLogEntry{
			Time:         getStringField(ev, "TimeCreated"),
			Level:        level,
			ProviderName: provider,
			EventID:      getIntField(ev, "Id"),
			Message:      truncateStr(getStringField(ev, "Message"), 200),
		}

		if len(summary.RecentEvents) < 50 {
			summary.RecentEvents = append(summary.RecentEvents, entry)
		}

		if level == "Error" || level == "错误" || level == "Critical" || level == "严重" {
			if len(summary.RecentErrors) < 20 {
				summary.RecentErrors = append(summary.RecentErrors, entry)
			}
		}
	}

	// Sort providers by count and take top 10
	type kv struct {
		k string
		v int
	}
	var sorted []kv
	for k, v := range providerCount {
		sorted = append(sorted, kv{k, v})
	}
	// Simple bubble sort for small list
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].v > sorted[i].v {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	limit := 10
	if len(sorted) < limit {
		limit = len(sorted)
	}
	for i := 0; i < limit; i++ {
		summary.TopProviders = append(summary.TopProviders, EventProvider{
			Name:  sorted[i].k,
			Count: sorted[i].v,
		})
	}
}

func getStringField(m map[string]interface{}, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return fmt.Sprintf("%v", v)
	}
	return s
}

func getIntField(m map[string]interface{}, key string) int {
	v, ok := m[key]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case string:
		id, _ := strconv.Atoi(n)
		return id
	}
	return 0
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// tryWevtUtilFallback queries event count and critical/error/warning counts via wevtutil
func tryWevtUtilFallback(hoursBack int, maxEvents int, summary EventLogSummary) EventLogSummary {
	// Get total count
	countCmd := exec.Command("wevtutil", "qe", "System", "/c:1", "/rd:true", "/f:text")
	countOut, _ := countCmd.CombinedOutput()

	reCount := regexp.MustCompile(`Event\[(\d+)\]`)
	if matches := reCount.FindStringSubmatch(string(countOut)); len(matches) > 1 {
		// This won't give real total, but we can count via different approach
		_ = matches
	}

	// Use wevtutil to query recent events in text format
	cmd := exec.Command("wevtutil", "qe", "System",
		"/c:"+strconv.Itoa(maxEvents),
		"/rd:true",
		"/f:text",
		"/e:root")
	output, err := cmd.CombinedOutput()
	if err != nil {
		summary.RawOutput += fmt.Sprintf("\nwevtutil also failed: %v", err)
		return summary
	}

	text := string(output)
	summary.RawOutput = text

	// Count levels from text output
	summary.CriticalCount = strings.Count(strings.ToLower(text), "level: 1")
	summary.ErrorCount = strings.Count(strings.ToLower(text), "level: 2")
	summary.WarningCount = strings.Count(strings.ToLower(text), "level: 3")
	summary.InfoCount = strings.Count(strings.ToLower(text), "level: 4")

	// Estimate total
	summary.TotalEvents = summary.CriticalCount + summary.ErrorCount + summary.WarningCount + summary.InfoCount

	return summary
}

// enrichWithWevtUtil runs a quick wevtutil count to get more accurate summary
func enrichWithWevtUtil(summary *EventLogSummary) {
	// Quick count via PowerShell for critical/error/warning
	ps := `$c=0;$e=0;$w=0;$i=0;` +
		`Get-WinEvent -LogName System -MaxEvents 1000 | ForEach-Object {` +
		`switch($_.LevelDisplayName){'Critical'{$c++}'Error'{$e++}'Warning'{$w++}default{$i++}}};` +
		`"Critical:$c Error:$e Warning:$w Info:$i"`
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", ps)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return
	}

	re := regexp.MustCompile(`Critical:(\d+) Error:(\d+) Warning:(\d+) Info:(\d+)`)
	matches := re.FindStringSubmatch(string(out))
	if len(matches) == 5 {
		c, _ := strconv.Atoi(matches[1])
		e, _ := strconv.Atoi(matches[2])
		w, _ := strconv.Atoi(matches[3])
		inf, _ := strconv.Atoi(matches[4])
		summary.CriticalCount = c
		summary.ErrorCount = e
		summary.WarningCount = w
		summary.InfoCount = inf
		summary.TotalEvents = c + e + w + inf
	}
}

// ExportSystemEventLog exports System event log to a CSV file
func ExportSystemEventLog(outputPath string, hoursBack int, maxEvents int) string {
	if outputPath == "" {
		outputPath = `C:\Users\Public\ETL_Traces\SystemEventLog.csv`
	}
	if maxEvents <= 0 {
		maxEvents = 1000
	}

	var psScript string
	if hoursBack > 0 {
		psScript = fmt.Sprintf(
			`$start = (Get-Date).AddHours(-%d); `+
				`Get-WinEvent -LogName System -MaxEvents %d | `+
				`Where-Object { $_.TimeCreated -ge $start } | `+
				`Select-Object TimeCreated, LevelDisplayName, ProviderName, Id, @{n='Message';e={($_.Message -replace '\s+',' ').Trim()}} | `+
				`Export-Csv -Path '%s' -NoTypeInformation -Encoding UTF8`,
			hoursBack, maxEvents, outputPath)
	} else {
		psScript = fmt.Sprintf(
			`Get-WinEvent -LogName System -MaxEvents %d | `+
				`Select-Object TimeCreated, LevelDisplayName, ProviderName, Id, @{n='Message';e={($_.Message -replace '\s+',' ').Trim()}} | `+
				`Export-Csv -Path '%s' -NoTypeInformation -Encoding UTF8`,
			maxEvents, outputPath)
	}

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Export failed: %v - %s", err, string(out))
	}
	return outputPath
}
