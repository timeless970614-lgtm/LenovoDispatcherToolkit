//go:build windows

package backend

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
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
	EventData     string `json:"eventData"`
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
				`$events | Select-Object TimeCreated, LevelDisplayName, ProviderName, Id, Message, `+
				`@{N='EventData';E={$x=$_.ToXml();[xml]$xml=$x;($xml.Event.EventData.Data | ForEach-Object { '{0}={1}' -f $_.Name, $_.'#text' }) -join [char]10}} | ConvertTo-Json -Depth 2`,
			hoursBack, maxEvents*10)
		summary.TimeRange = fmt.Sprintf("Last %d hours (up to %d events)", hoursBack, maxEvents)
	} else {
			psScript = fmt.Sprintf(
			`Get-WinEvent -LogName System -MaxEvents %d | `+
				`Select-Object TimeCreated, LevelDisplayName, ProviderName, Id, Message, `+
				`@{N='EventData';E={$x=$_.ToXml();[xml]$xml=$x;($xml.Event.EventData.Data | ForEach-Object { '{0}={1}' -f $_.Name, $_.'#text' }) -join [char]10}} | ConvertTo-Json -Depth 2`,
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
		case "Critical", "ä¸¥é‡":
			summary.CriticalCount++
		case "Error", "é”™è¯¯":
			summary.ErrorCount++
		case "Warning", "è­¦å‘Š":
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
			Message:      truncateStr(getStringField(ev, "Message"), 500),
			EventData:    truncateStr(getStringField(ev, "EventData"), 2000),
		}

		if len(summary.RecentEvents) < 50 {
			summary.RecentEvents = append(summary.RecentEvents, entry)
		}

		if level == "Error" || level == "é”™è¯¯" || level == "Critical" || level == "ä¸¥é‡" {
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
	countOut, err := countCmd.CombinedOutput()

	if err != nil {
		// Count query failed, keep countOut nil
	}

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
		c, err := strconv.Atoi(matches[1])
		if err != nil {
			c = 0
		}
		e, err := strconv.Atoi(matches[2])
		if err != nil {
			e = 0
		}
		w, err := strconv.Atoi(matches[3])
		if err != nil {
			w = 0
		}
		inf, err := strconv.Atoi(matches[4])
		if err != nil {
			inf = 0
		}
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

// EventLogAnalysis holds the analysis result of system event log
type EventLogAnalysis struct {
	OverallHealth  string            `json:"overallHealth"`  // Good / Warning / Critical
	Summary        string            `json:"summary"`
	RepeatErrors   []RepeatPattern   `json:"repeatErrors"`
	RootCauses     []RootCause       `json:"rootCauses"`
	Correlations   []Correlation     `json:"correlations"`
	KnownIssues    []KnownIssue      `json:"knownIssues"`
	Recommendations []string         `json:"recommendations"`
	Timeline       []TimelineEvent    `json:"timeline"`
}

type RepeatPattern struct {
	ProviderName string `json:"providerName"`
	EventID      int    `json:"eventId"`
	Count        int    `json:"count"`
	LastSeen     string `json:"lastSeen"`
	SampleMsg    string `json:"sampleMsg"`
}

type RootCause struct {
	Category string `json:"category"`
	Detail  string `json:"detail"`
	Events  int    `json:"events"`
	Severity string `json:"severity"`
}

type Correlation struct {
	Description string `json:"description"`
	EventIDs    []int  `json:"eventIds"`
	Timestamps  []string `json:"timestamps"`
}

type KnownIssue struct {
	Name       string `json:"name"`
	Matched    bool   `json:"matched"`
	Confidence int    `json:"confidence"` // 0-100
	Detail     string `json:"detail"`
}

type TimelineEvent struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	Source  string `json:"source"`
	ID      int    `json:"id"`
	Message string `json:"message"`
}

// AnalyzeSystemEventLog performs pattern analysis on captured event log data
func AnalyzeSystemEventLog(summary EventLogSummary) EventLogAnalysis {
	analysis := EventLogAnalysis{
		RepeatErrors:   []RepeatPattern{},
		RootCauses:     []RootCause{},
		Correlations:   []Correlation{},
		KnownIssues:    []KnownIssue{},
		Recommendations: []string{},
		Timeline:       []TimelineEvent{},
	}

	// 1. Overall health assessment
	total := summary.TotalEvents
	crit := summary.CriticalCount
	errs := summary.ErrorCount
	warns := summary.WarningCount
	if total == 0 {
		analysis.OverallHealth = "Good"
		analysis.Summary = "No System events captured in the selected time range."
		return analysis
	}

	if crit > 5 {
		analysis.OverallHealth = "Critical"
		analysis.Summary = fmt.Sprintf("System health is CRITICAL: %d critical and %d error events detected out of %d total events. Immediate investigation recommended.", crit, errs, total)
	} else if errs > 20 {
		analysis.OverallHealth = "Warning"
		analysis.Summary = fmt.Sprintf("System health is WARNING: %d error events and %d warnings detected. Some issues need attention.", errs, warns)
	} else if errs > 5 || warns > 50 {
		analysis.OverallHealth = "Warning"
		analysis.Summary = fmt.Sprintf("System health is WARNING: %d errors and %d warnings in the selected time range.", errs, warns)
	} else {
		analysis.OverallHealth = "Good"
		analysis.Summary = fmt.Sprintf("System health is GOOD: %d events captured, only %d errors and %d warnings â€” within normal range.", total, errs, warns)
	}

	// 2. Find repeat error patterns
	eventCount := make(map[string]int) // "ProviderName:EventID" -> count
	eventSamples := make(map[string]string)
	eventLastSeen := make(map[string]string)
	eventProviders := make(map[string]string)
	eventIDs := make(map[string]int)

	for _, ev := range summary.RecentEvents {
		key := fmt.Sprintf("%s:%d", ev.ProviderName, ev.EventID)
		eventCount[key]++
		if _, ok := eventSamples[key]; !ok && ev.Message != "" {
			eventSamples[key] = ev.Message
		}
		if ev.Time != "" {
			eventLastSeen[key] = ev.Time
		}
		eventProviders[key] = ev.ProviderName
		eventIDs[key] = ev.EventID
	}

	// Sort by count descending
	type kv struct {
		key   string
		count int
	}
	var sorted []kv
	for k, v := range eventCount {
		sorted = append(sorted, kv{k, v})
	}
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].count > sorted[i].count {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	for i, s := range sorted {
		if i >= 10 {
			break
		}
		if s.count >= 2 {
			analysis.RepeatErrors = append(analysis.RepeatErrors, RepeatPattern{
				ProviderName: eventProviders[s.key],
				EventID:      eventIDs[s.key],
				Count:        s.count,
				LastSeen:     eventLastSeen[s.key],
				SampleMsg:    truncateStr(eventSamples[s.key], 150),
			})
		}
	}

	// 3. Root cause analysis based on known event patterns
	if crit > 0 {
		analysis.RootCauses = append(analysis.RootCauses, RootCause{
			Category: "Critical Events",
			Detail:  fmt.Sprintf("%d critical events detected. These indicate system failures or unexpected shutdowns.", crit),
			Events:  crit,
			Severity: "Critical",
		})
	}

	// Check for driver-related errors
	driverErrors := 0
	for _, ev := range summary.RecentEvents {
		if (ev.Level == "Error" || ev.Level == "Critical") && ev.EventID != 0 {
			msg := strings.ToLower(ev.Message)
			prov := strings.ToLower(ev.ProviderName)
			if strings.Contains(msg, "driver") || strings.Contains(prov, "driver") ||
				strings.Contains(msg, "device") || strings.Contains(prov, "wudf") ||
				strings.Contains(prov, "ndis") || strings.Contains(prov, "tcpip") {
				driverErrors++
			}
		}
	}
	if driverErrors > 0 {
		analysis.RootCauses = append(analysis.RootCauses, RootCause{
			Category: "Driver Issues",
			Detail:  fmt.Sprintf("%d driver/device related errors found. Check Device Manager for hardware conflicts or outdated drivers.", driverErrors),
			Events:  driverErrors,
			Severity: "Warning",
		})
	}

	// Check for service crashes
	serviceErrors := 0
	for _, ev := range summary.RecentEvents {
		if ev.EventID == 7031 || ev.EventID == 7032 || ev.EventID == 7034 || ev.EventID == 7000 || ev.EventID == 7001 || ev.EventID == 7002 || ev.EventID == 7009 {
			serviceErrors++
		}
	}
	if serviceErrors > 0 {
		analysis.RootCauses = append(analysis.RootCauses, RootCause{
			Category: "Service Failures",
			Detail:  fmt.Sprintf("%d service-related events (crash/start/stop failures). Affected services may need investigation.", serviceErrors),
			Events:  serviceErrors,
			Severity: func() string { if serviceErrors > 5 { return "Critical" }; return "Warning" }(),
		})
	}

	// Check for disk/filesystem issues
	diskIssues := 0
	for _, ev := range summary.RecentEvents {
		if ev.EventID == 7 || ev.EventID == 9 || ev.EventID == 11 || ev.EventID == 15 || ev.EventID == 51 || ev.EventID == 153 {
			diskIssues++
		}
		msg := strings.ToLower(ev.Message)
		if strings.Contains(msg, "disk") || strings.Contains(msg, "storage") || strings.Contains(msg, "ntfs") {
			diskIssues++
		}
	}
	if diskIssues > 0 {
		analysis.RootCauses = append(analysis.RootCauses, RootCause{
			Category: "Disk/Filesystem",
			Detail:  fmt.Sprintf("%d disk or filesystem events detected. Run chkdsk if disk errors persist.", diskIssues),
			Events:  diskIssues,
			Severity: "Warning",
		})
	}

	// Check for network issues
	netEvents := 0
	for _, ev := range summary.RecentEvents {
		prov := strings.ToLower(ev.ProviderName)
		if strings.Contains(prov, "net") || strings.Contains(prov, "wlan") || strings.Contains(prov, "dhcp") || strings.Contains(prov, "dns") {
			if ev.Level == "Error" || ev.Level == "Warning" {
				netEvents++
			}
		}
	}
	if netEvents > 0 {
		analysis.RootCauses = append(analysis.RootCauses, RootCause{
			Category: "Network Issues",
			Detail:  fmt.Sprintf("%d network-related warnings/errors. Check adapter settings and connectivity.", netEvents),
			Events:  netEvents,
			Severity: "Info",
		})
	}

	// Check for thermal/power issues
	thermEvents := 0
	for _, ev := range summary.RecentEvents {
		prov := strings.ToLower(ev.ProviderName)
		msg := strings.ToLower(ev.Message)
		if strings.Contains(prov, "acpi") || strings.Contains(msg, "thermal") || strings.Contains(msg, "power") ||
			strings.Contains(msg, "battery") || strings.Contains(msg, "surge") {
			thermEvents++
		}
	}
	if thermEvents > 0 {
		analysis.RootCauses = append(analysis.RootCauses, RootCause{
			Category: "Thermal/Power",
			Detail:  fmt.Sprintf("%d thermal or power management events. May indicate cooling issues or power supply problems.", thermEvents),
			Events:  thermEvents,
			Severity: "Info",
		})
	}

	// 4. Known issue patterns
	knownPatterns := []struct {
		name        string
		eventIDs    []int
		providers   []string
		keywords    []string
		confidence  int
		detail      string
	}{
		{"Unexpected Shutdown (Kernel-Power 41)", []int{41}, []string{"Kernel-Power"}, nil, 95, "The system went to sleep/hibernate but encountered an issue, or lost power unexpectedly. Check power settings and hardware."},
		{"BSOD (BugCheck)", []int{1001}, []string{"BugCheck", "WerFault"}, nil, 90, "A Blue Screen of Death occurred. Check minidump files in C:\\Windows\\Minidump for crash analysis."},
		{"Driver Verifier Detected Violation", []int{0xC4, 0xC1, 0xC9, 0xC6, 0x10A}, nil, nil, 80, "Driver Verifier caught a driver violating rules. The violating driver name is usually in the crash message."},
		{"Windows Update Failure", []int{20}, []string{"WindowsUpdateClient"}, nil, 85, "Windows Update encountered an error. Run Windows Update Troubleshooter or check update history."},
		{"DNS Client Issues", []int{1014}, []string{"DNS Client Events"}, nil, 75, "DNS resolution failures detected. Check DNS server settings or try flushing DNS cache."},
		{"Application Hang (Application Error 1002)", []int{1002}, []string{"Application Error"}, nil, 70, "An application became unresponsive. The program name is in the event details."},
		{"Print Spooler Issues", []int{}, []string{"PrintService"}, []string{"spooler"}, 70, "Print spooler service encountered errors. Restarting the spooler service may help."},
		{"USB Device Surprises", []int{}, nil, []string{"surprise", "usb"}, 65, "USB device was unexpectedly removed. Check USB connections and device health."},
	}

	for _, kp := range knownPatterns {
		matched := false
		matchCount := 0
		for _, ev := range summary.RecentEvents {
			// Check event ID
			for _, eid := range kp.eventIDs {
				if ev.EventID == eid {
					matched = true
					matchCount++
				}
			}
			// Check provider
			for _, prov := range kp.providers {
				if ev.ProviderName == prov {
					matched = true
					matchCount++
				}
			}
			// Check keywords in message
			if len(kp.keywords) > 0 {
				msg := strings.ToLower(ev.Message)
				for _, kw := range kp.keywords {
					if strings.Contains(msg, strings.ToLower(kw)) {
						matched = true
						matchCount++
					}
				}
			}
		}
		if matched {
			analysis.KnownIssues = append(analysis.KnownIssues, KnownIssue{
				Name:       kp.name,
				Matched:    true,
				Confidence: min(kp.confidence+matchCount*5, 100),
				Detail:     kp.detail,
			})
		}
	}

	// 5. Correlations - find event bursts
	if len(summary.RecentEvents) > 5 {
		type timeBucket struct {
			timeWindow string
			events      int
		}
		buckets := []timeBucket{}
		for _, ev := range summary.RecentEvents {
			if ev.Time == "" {
				continue
			}
			// Extract hour:minute bucket
			bucket := ""
			re := regexp.MustCompile(`(\d{2}:\d{2})`)
			m := re.FindStringSubmatch(ev.Time)
			if len(m) > 1 {
				bucket = m[1]
			} else {
				continue
			}
			found := false
			for idx, b := range buckets {
				if b.timeWindow == bucket {
					buckets[idx].events++
					found = true
					break
				}
			}
			if !found {
				buckets = append(buckets, timeBucket{timeWindow: bucket, events: 1})
			}
		}
		// Find burst windows (more than 3x average)
		if len(buckets) > 1 {
			avg := 0
			for _, b := range buckets {
				avg += b.events
			}
			avg = avg / len(buckets)
			for _, b := range buckets {
				if b.events > avg*3 && b.events > 10 {
					analysis.Correlations = append(analysis.Correlations, Correlation{
						Description: fmt.Sprintf("Event burst at %s: %d events in 1 minute (%dx above average of %d/minute)", b.timeWindow, b.events, b.events/avg, avg),
					})
				}
			}
		}
	}

	// 6. Timeline - build chronologically sorted event list for errors/critical only
	for _, ev := range summary.RecentEvents {
		if ev.Level == "Critical" || ev.Level == "Error" {
			analysis.Timeline = append(analysis.Timeline, TimelineEvent{
				Time:    ev.Time,
				Level:   ev.Level,
				Source:  ev.ProviderName,
				ID:      ev.EventID,
				Message: truncateStr(ev.Message, 120),
			})
		}
	}
	// Reverse to chronological order (oldest first)
	for i, j := 0, len(analysis.Timeline)-1; i < j; i, j = i+1, j-1 {
		analysis.Timeline[i], analysis.Timeline[j] = analysis.Timeline[j], analysis.Timeline[i]
	}
	if len(analysis.Timeline) > 20 {
		analysis.Timeline = analysis.Timeline[:20]
	}

	// 7. Recommendations
	if crit > 0 {
		analysis.Recommendations = append(analysis.Recommendations, "Critical events detected â€” check minidump files (C:\\Windows\\Minidump) and run system file checker (sfc /scannow).")
	}
	if driverErrors > 0 {
		analysis.Recommendations = append(analysis.Recommendations, "Driver issues found â€” open Device Manager and check for yellow warning icons. Consider updating problematic drivers.")
	}
	if serviceErrors > 0 {
		analysis.Recommendations = append(analysis.Recommendations, "Service failures detected â€” check Windows Services console for services with abnormal restart behavior.")
	}
	if diskIssues > 0 {
		analysis.Recommendations = append(analysis.Recommendations, "Disk issues detected â€” run chkdsk /f on affected drives and check SMART health status.")
	}
	if netEvents > 3 {
		analysis.Recommendations = append(analysis.Recommendations, "Network issues found â€” verify Wi-Fi/Ethernet adapter settings and check for firmware updates.")
	}
	if thermEvents > 0 {
		analysis.Recommendations = append(analysis.Recommendations, "Thermal/power events detected â€” monitor temperatures and check power supply health if recurring.")
	}
	if len(analysis.RepeatErrors) > 3 {
		analysis.Recommendations = append(analysis.Recommendations, fmt.Sprintf("Multiple repeating errors found (%d patterns). Focus on the most frequent one first.", len(analysis.RepeatErrors)))
	}
	if len(analysis.Recommendations) == 0 {
		analysis.Recommendations = append(analysis.Recommendations, "System appears healthy. No immediate action required.")
	}

	return analysis
}


// ============ EVTX File Loading & Conversion ============

// LoadEVTXFile parses an .evtx file and returns EventLogSummary
func LoadEVTXFile(evtxPath string) EventLogSummary {
	summary := EventLogSummary{
		TimeRange: "File: " + evtxPath,
	}
	psScript := fmt.Sprintf(
		"$events = Get-WinEvent -Path '%s' -MaxEvents 2000 -ErrorAction Stop; "+
			"$events | Select-Object TimeCreated, LevelDisplayName, ProviderName, Id, Message, "+
			"@{N='EventData';E={$x=$_.ToXml();[xml]$xml=$x;($xml.Event.EventData.Data | ForEach-Object { '{0}={1}' -f $_.Name, $_.'#text' }) -join [char]10}} | ConvertTo-Json -Depth 2",
		evtxPath)
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
	output, err := cmd.CombinedOutput()
	if err != nil {
		summary.RawOutput = fmt.Sprintf("Failed to read EVTX: %v", err)
		return summary
	}
	rawJSON := strings.TrimSpace(string(output))
	summary.RawOutput = rawJSON
	parseEventLogJSON(rawJSON, &summary)
	if summary.TotalEvents > 0 && len(summary.RecentEvents) > 0 {
		first := summary.RecentEvents[0].Time
		last := summary.RecentEvents[len(summary.RecentEvents)-1].Time
		summary.TimeRange = fmt.Sprintf("%d events (%s ~ %s)", summary.TotalEvents, first, last)
	}
	return summary
}

// ExportEVTXToCSV converts an .evtx file to CSV format
func ExportEVTXToCSV(evtxPath string, outputPath string) string {
	if outputPath == "" {
		outputPath = strings.TrimSuffix(evtxPath, filepath.Ext(evtxPath)) + ".csv"
	}
	psScript := fmt.Sprintf(
		"Get-WinEvent -Path '%s' -MaxEvents 5000 -ErrorAction Stop | "+
			"Select-Object TimeCreated, LevelDisplayName, Level, ProviderName, Id, "+
			"@{n='Message';e={($_.Message -replace '\\s+',' ').Trim()}} | "+
			"Export-Csv -Path '%s' -NoTypeInformation -Encoding UTF8",
		evtxPath, outputPath)
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
	_, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Export failed: %v", err)
	}
	return outputPath
}

// GetEVTXEventsByLevel returns events from an .evtx file filtered by level
func GetEVTXEventsByLevel(evtxPath string, level string, maxEvents int) []EventLogEntry {
	if maxEvents <= 0 {
		maxEvents = 500
	}
	var psScript string
	if level != "" {
		psScript = fmt.Sprintf(
			"Get-WinEvent -Path '%s' -MaxEvents %d -ErrorAction Stop | "+
				"Where-Object { $_.LevelDisplayName -eq '%s' } | "+
				"Select-Object TimeCreated, LevelDisplayName, ProviderName, Id, Message, "+
				"@{N='EventData';E={$x=$_.ToXml();[xml]$xml=$x;($xml.Event.EventData.Data | ForEach-Object { '{0}={1}' -f $_.Name, $_.'#text' }) -join [char]10}} | ConvertTo-Json -Depth 2",
			evtxPath, maxEvents*5, level)
	} else {
		psScript = fmt.Sprintf(
			"Get-WinEvent -Path '%s' -MaxEvents %d -ErrorAction Stop | "+
				"Select-Object TimeCreated, LevelDisplayName, ProviderName, Id, Message, "+
				"@{N='EventData';E={$x=$_.ToXml();[xml]$xml=$x;($xml.Event.EventData.Data | ForEach-Object { '{0}={1}' -f $_.Name, $_.'#text' }) -join [char]10}} | ConvertTo-Json -Depth 2",
			evtxPath, maxEvents)
	}
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil
	}
	rawJSON := strings.TrimSpace(string(output))
	raw := strings.TrimSpace(rawJSON)
	if !strings.HasPrefix(raw, "[") {
		raw = "[" + raw + "]"
	}
	var events []map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &events); err != nil {
		return nil
	}
	var entries []EventLogEntry
	for _, ev := range events {
		entry := EventLogEntry{
			Time:         getStringField(ev, "TimeCreated"),
			Level:        getStringField(ev, "LevelDisplayName"),
			ProviderName: getStringField(ev, "ProviderName"),
			EventID:      getIntField(ev, "Id"),
			Message:      truncateStr(getStringField(ev, "Message"), 200),
			EventData:    truncateStr(getStringField(ev, "EventData"), 2000),
		}
		entries = append(entries, entry)
		if len(entries) >= maxEvents {
			break
		}
	}
	return entries
}
