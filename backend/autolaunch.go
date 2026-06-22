//go:build windows

package backend

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

// AutoLaunchItem represents a single launchable item
type AutoLaunchItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`    // "browser", "app", "protocol", "folder"
	LaunchType  string `json:"launchType"`  // "url", "exe", "protocol", "folder"
	LaunchValue string `json:"launchValue"` // URL / exe path / protocol URI / folder path
	WaitSec     int    `json:"waitSec"`     // Seconds to wait after launch (0 = no wait)
	Enabled     bool   `json:"enabled"`     // Default enabled state
	Description string `json:"description"` // Brief description
}

// AutoLaunchResult represents the result of launching a single item
type AutoLaunchResult struct {
	ItemID  string `json:"itemId"`
	Name    string `json:"name"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	WaitSec int    `json:"waitSec"`
}

// AutoLaunchFolderConfig holds folder-scanning config
type AutoLaunchFolderConfig struct {
	Path      string   `json:"path"`      // Folder to scan
	Excludes  []string `json:"excludes"`  // Filenames to skip
	WaitSec   int      `json:"waitSec"`   // Wait between each file open
	Enabled   bool     `json:"enabled"`   // Whether to include folder items
}

var (
	// autoLaunchItems defines the built-in launch items (inspired by open_all_files.bat)
	autoLaunchItems = []AutoLaunchItem{
		// === Browser tabs (Edge) ===
		{
			ID: "edge-blank", Name: "Edge 浏览器", Category: "browser",
			LaunchType: "url", LaunchValue: "microsoft-edge:",
			WaitSec: 30, Enabled: true, Description: "打开 Edge 浏览器空白页",
		},
		{
			ID: "edge-weibo", Name: "微博", Category: "browser",
			LaunchType: "url", LaunchValue: "microsoft-edge:https://weibo.com",
			WaitSec: 30, Enabled: true, Description: "通过 Edge 打开微博",
		},
		{
			ID: "edge-jd", Name: "京东", Category: "browser",
			LaunchType: "url", LaunchValue: "microsoft-edge:https://www.jd.com",
			WaitSec: 10, Enabled: true, Description: "通过 Edge 打开京东",
		},
		{
			ID: "edge-qqmail", Name: "QQ 邮箱", Category: "browser",
			LaunchType: "url", LaunchValue: "microsoft-edge:https://mail.qq.com",
			WaitSec: 10, Enabled: true, Description: "通过 Edge 打开 QQ 邮箱",
		},
		{
			ID: "edge-iqiyi", Name: "爱奇艺 1080p", Category: "browser",
			LaunchType: "url", LaunchValue: "microsoft-edge:https://www.iqiyi.com/v_19rr8av3a8.html",
			WaitSec: 10, Enabled: true, Description: "通过 Edge 播放爱奇艺 1080p 视频",
		},

		// === Protocol launches ===
		{
			ID: "outlook", Name: "Outlook", Category: "protocol",
			LaunchType: "protocol", LaunchValue: "outlook:",
			WaitSec: 10, Enabled: true, Description: "打开 Outlook 邮件客户端",
		},

		// === Local apps ===
		{
			ID: "iqiyi-app", Name: "爱奇艺客户端", Category: "app",
			LaunchType: "exe", LaunchValue: `C:\Program Files\IQIYI Video\LStyle\QyClient.exe`,
			WaitSec: 15, Enabled: true, Description: "打开爱奇艺桌面应用播放 1080p 视频",
		},
	}

	// Default folder config: scan the bat file's own directory
	defaultFolderConfig = AutoLaunchFolderConfig{
		Path:     "", // Empty = use the exe's directory at runtime
		Excludes: []string{"open_all_files.bat", "always_on_top.ps1"},
		WaitSec:  30,
		Enabled:  true,
	}

	autoLaunchMu sync.Mutex
)

// GetAutoLaunchItems returns the built-in launch items plus folder items
func GetAutoLaunchItems() ([]AutoLaunchItem, AutoLaunchFolderConfig) {
	return autoLaunchItems, defaultFolderConfig
}

// SetAutoLaunchFolderConfig updates the folder scanning config
func SetAutoLaunchFolderConfig(cfg AutoLaunchFolderConfig) {
	autoLaunchMu.Lock()
	defer autoLaunchMu.Unlock()
	defaultFolderConfig = cfg
}

// LaunchAutoLaunchItem launches a single item by its ID
func LaunchAutoLaunchItem(itemID string) AutoLaunchResult {
	for _, item := range autoLaunchItems {
		if item.ID == itemID {
			return launchItem(item)
		}
	}
	// Check folder items
	return AutoLaunchResult{
		ItemID: itemID,
		Name:   itemID,
		Error:  "item not found",
	}
}

// BatchLaunchAutoLaunchItems launches multiple items sequentially
func BatchLaunchAutoLaunchItems(itemIDs []string) []AutoLaunchResult {
	var results []AutoLaunchResult
	for _, id := range itemIDs {
		result := LaunchAutoLaunchItem(id)
		results = append(results, result)
		if result.WaitSec > 0 && result.Success {
			time.Sleep(time.Duration(result.WaitSec) * time.Second)
		}
	}
	return results
}

// LaunchAllEnabledItems launches all enabled items (built-in + folder)
func LaunchAllEnabledItems() []AutoLaunchResult {
	var results []AutoLaunchResult

	// Launch built-in items
	for _, item := range autoLaunchItems {
		if item.Enabled {
			result := launchItem(item)
			results = append(results, result)
			if result.WaitSec > 0 && result.Success {
				time.Sleep(time.Duration(result.WaitSec) * time.Second)
			}
		}
	}

	// Launch folder items
	if defaultFolderConfig.Enabled {
		folderResults := launchFolderItems()
		results = append(results, folderResults...)
	}

	return results
}

// GetFolderFiles returns the list of files in the configured folder
func GetFolderFiles() ([]AutoLaunchItem, error) {
	autoLaunchMu.Lock()
	folderPath := defaultFolderConfig.Path
	excludes := defaultFolderConfig.Excludes
	autoLaunchMu.Unlock()

	if folderPath == "" {
		exePath, err := os.Executable()
		if err != nil {
			return nil, fmt.Errorf("cannot determine exe path: %w", err)
		}
		folderPath = filepath.Dir(exePath)
	}

	var items []AutoLaunchItem
	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read folder %s: %w", folderPath, err)
	}

	excludeSet := make(map[string]bool)
	for _, e := range excludes {
		excludeSet[strings.ToLower(e)] = true
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if excludeSet[strings.ToLower(name)] {
			continue
		}
		// Skip .bat and .ps1 files that are scripts
		ext := strings.ToLower(filepath.Ext(name))
		if ext == ".bat" || ext == ".ps1" || ext == ".cmd" {
			continue
		}

		fullPath := filepath.Join(folderPath, name)
		items = append(items, AutoLaunchItem{
			ID:          "folder-" + name,
			Name:        name,
			Category:    "folder",
			LaunchType:  "folder",
			LaunchValue: fullPath,
			WaitSec:     defaultFolderConfig.WaitSec,
			Enabled:     true,
			Description: "文件夹文件: " + fullPath,
		})
	}

	return items, nil
}

// launchItem executes a single launch item
func launchItem(item AutoLaunchItem) AutoLaunchResult {
	result := AutoLaunchResult{
		ItemID:  item.ID,
		Name:    item.Name,
		WaitSec: item.WaitSec,
	}

	var cmd *exec.Cmd
	switch item.LaunchType {
	case "url":
		// Edge protocol URLs like "microsoft-edge:https://..."
		cmd = exec.Command("cmd", "/c", "start", "", item.LaunchValue)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	case "exe":
		// Check if exe exists first
		if _, err := os.Stat(item.LaunchValue); os.IsNotExist(err) {
			result.Success = false
			result.Error = fmt.Sprintf("文件不存在: %s", item.LaunchValue)
			return result
		}
		cmd = visibleCmd(item.LaunchValue)

	case "protocol":
		// URI protocol like "outlook:"
		cmd = exec.Command("cmd", "/c", "start", "", item.LaunchValue)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	case "folder":
		// Open a file from folder
		if _, err := os.Stat(item.LaunchValue); os.IsNotExist(err) {
			result.Success = false
			result.Error = fmt.Sprintf("文件不存在: %s", item.LaunchValue)
			return result
		}
		cmd = exec.Command("cmd", "/c", "start", "", "", item.LaunchValue)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	default:
		result.Success = false
		result.Error = "unknown launch type: " + item.LaunchType
		return result
	}

	if err := cmd.Start(); err != nil {
		result.Success = false
		result.Error = err.Error()
		return result
	}

	result.Success = true
	return result
}

// launchFolderItems opens all files in the configured folder
func launchFolderItems() []AutoLaunchResult {
	items, err := GetFolderFiles()
	if err != nil {
		return []AutoLaunchResult{{
			ItemID: "folder-error",
			Name:   "文件夹扫描",
			Error:  err.Error(),
		}}
	}

	var results []AutoLaunchResult
	for _, item := range items {
		if !item.Enabled {
			continue
		}
		result := launchItem(item)
		results = append(results, result)
		if item.WaitSec > 0 && result.Success {
			time.Sleep(time.Duration(item.WaitSec) * time.Second)
		}
	}
	return results
}

// ToggleAutoLaunchItem enables/disables a built-in item
func ToggleAutoLaunchItem(itemID string, enabled bool) {
	for i := range autoLaunchItems {
		if autoLaunchItems[i].ID == itemID {
			autoLaunchItems[i].Enabled = enabled
			return
		}
	}
}

// SetAutoLaunchItemWait updates the wait time for a built-in item
func SetAutoLaunchItemWait(itemID string, waitSec int) {
	for i := range autoLaunchItems {
		if autoLaunchItems[i].ID == itemID {
			autoLaunchItems[i].WaitSec = waitSec
			return
		}
	}
}
