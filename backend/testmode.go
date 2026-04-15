package backend

import (
	"os/exec"
	"strings"
)

// OpenTestMode enables Windows test mode for driver testing
// This allows loading unsigned/test-signed drivers
func OpenTestMode() map[string]interface{} {
	// Execute bcdedit to enable test signing
	cmd := exec.Command("bcdedit", "/set", "testsigning", "on")
	output, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(output))

	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "执行失败: " + outputStr,
		}
	}

	// Check if operation was successful
	if strings.Contains(outputStr, "操作成功完成") || strings.Contains(strings.ToLower(outputStr), "success") {
		return map[string]interface{}{
			"success": true,
			"message": "测试模式已启用！请重启电脑生效。\n\n桌面右下角将显示'测试模式'水印。",
		}
	}

	return map[string]interface{}{
		"success": false,
		"message": "未知结果: " + outputStr,
	}
}
