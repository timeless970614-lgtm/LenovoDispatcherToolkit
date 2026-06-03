//go:build windows

package backend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	NVIDIA_API_BASE   = "https://integrate.api.nvidia.com/v1"
	NVIDIA_CONFIG_FILE = "nvidia_api_config.json"
)

// NVIDIAChatMessage represents a chat completion message
type NVIDIAChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// NVIDIAChatRequest represents the API request body
type NVIDIAChatRequest struct {
	Model       string              `json:"model"`
	Messages    []NVIDIAChatMessage `json:"messages"`
	MaxTokens   int                 `json:"max_tokens,omitempty"`
	Temperature float64             `json:"temperature,omitempty"`
	Stream      bool                `json:"stream,omitempty"`
}

// NVIDIAChatResponse represents the API response
type NVIDIAChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Choices []struct {
		Index        int               `json:"index"`
		Message      NVIDIAChatMessage `json:"message"`
		FinishReason string            `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

// NVIDIAAPIConfig stores the NVIDIA API configuration
type NVIDIAAPIConfig struct {
	APIKey  string `json:"apiKey"`
	Model   string `json:"model"`
	Enabled bool   `json:"enabled"`
	BaseURL string `json:"baseUrl"`
}

// NVIDIAModelItem represents a selectable model for the UI
type NVIDIAModelItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Desc string `json:"desc"`
	Cat  string `json:"cat"`
}

// DefaultModels returns the built-in model selection list
func DefaultModels() []NVIDIAModelItem {
	return []NVIDIAModelItem{
		{ID: "z-ai/glm-5.1", Name: "GLM-5.1", Desc: "旗舰 LLM，Agent 工作流/编程/长程推理", Cat: "Chat"},
		{ID: "deepseek-ai/deepseek-v4-flash", Name: "DeepSeek V4 Flash", Desc: "284B MoE，1M 上下文，快速编程/Agent", Cat: "Chat"},
		{ID: "minimaxai/minimax-m2.7", Name: "MiniMax M2.7", Desc: "230B 参数，编程/推理/办公", Cat: "Chat"},
		{ID: "stepfun-ai/step-3.7-flash", Name: "Step 3.7 Flash", Desc: "稀疏 MoE 多模态推理", Cat: "Chat"},
		{ID: "nvidia/nemotron-3-content-safety", Name: "Content Safety", Desc: "多语言多模态内容安全检测", Cat: "Safety"},
		{ID: "nvidia/nemotron-parse", Name: "Document Parse", Desc: "文档结构化解析", Cat: "Parse"},
	}
}

// GetNVIDIAConfigPath returns the config file path (next to exe)
func GetNVIDIAConfigPath() string {
	exe, _ := os.Executable()
	dir := filepath.Dir(exe)
	return filepath.Join(dir, NVIDIA_CONFIG_FILE)
}

// LoadNVIDIAConfig loads API config from JSON file
func LoadNVIDIAConfig() NVIDIAAPIConfig {
	cfg := NVIDIAAPIConfig{
		Model:   "z-ai/glm-5.1",
		BaseURL: NVIDIA_API_BASE,
	}
	data, err := os.ReadFile(GetNVIDIAConfigPath())
	if err != nil {
		return cfg
	}
	_ = json.Unmarshal(data, &cfg)
	if cfg.BaseURL == "" {
		cfg.BaseURL = NVIDIA_API_BASE
	}
	if cfg.Model == "" {
		cfg.Model = "z-ai/glm-5.1"
	}
	return cfg
}

// SaveNVIDIAConfig persists API config to JSON file
func SaveNVIDIAConfig(cfg NVIDIAAPIConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(GetNVIDIAConfigPath(), data, 0644)
}

// IsNVIDIAEnabled checks if the cloud API is configured and ready
func IsNVIDIAEnabled() bool {
	cfg := LoadNVIDIAConfig()
	return cfg.Enabled && strings.TrimSpace(cfg.APIKey) != ""
}

// CallNVIDIAChat sends a chat-completion request and returns the assistant reply.
// modelOverride lets the caller pick a one-off model; falls back to config if empty.
func CallNVIDIAChat(messages []NVIDIAChatMessage, modelOverride string) (string, error) {
	cfg := LoadNVIDIAConfig()
	if cfg.APIKey == "" {
		return "", fmt.Errorf("NVIDIA API Key 未配置")
	}

	model := cfg.Model
	if modelOverride != "" {
		model = modelOverride
	}
	if model == "" {
		model = "z-ai/glm-5.1"
	}

	reqBody := NVIDIAChatRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   4096,
		Temperature: 0.7,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("JSON 序列化失败: %v", err)
	}

	url := cfg.BaseURL + "/chat/completions"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败 (请检查网络): %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var chatResp NVIDIAChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}
	if chatResp.Error != nil {
		return "", fmt.Errorf("NVIDIA API [%s]: %s", chatResp.Error.Type, chatResp.Error.Message)
	}
	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("API 返回空结果 (HTTP %d)", resp.StatusCode)
	}

	result := chatResp.Choices[0].Message.Content
	if result == "" {
		result = "(AI 返回了空回复)"
	}
	return result, nil
}

// buildSystemPrompt creates the system context that tells the cloud LLM
// about the local machine and what it can do.
func buildSystemPrompt(infoJSON string) string {
	return fmt.Sprintf(
		`你是 Lenovo Dispatcher Toolkit 的 AI 系统助手。你可以分析用户的电脑状态并给出专业建议。

当前系统信息 (JSON):
%s

你的能力：
1. 分析系统性能指标（CPU/内存/GPU 使用率），判断是否异常并给出优化建议
2. 分析日志文件，定位错误原因，提供修复方案
3. 回答 Windows 系统相关技术问题
4. 根据系统状态推荐最优配置（电源计划、显卡模式等）
5. 分析 ETL Trace 文件内容
6. 生成 PowerShell 优化脚本

回答规则：
- 使用中文回答
- 回答简洁专业，重点突出
- 如有具体的优化操作，给出步骤
- 如果用户上传了文件，重点分析文件内容`, infoJSON)
}

// AskNVIDIACloud sends the user question (with system context) to NVIDIA API.
// This is the primary entry point called from the AI Agent UI.
func AskNVIDIACloud(question string) (string, error) {
	if !IsNVIDIAEnabled() {
		return "", fmt.Errorf("NVIDIA API 未启用")
	}

	info := GetAIAgentSystemInfo()
	infoJSON, _ := json.Marshal(info)
	sysPrompt := buildSystemPrompt(string(infoJSON))

	messages := []NVIDIAChatMessage{
		{Role: "system", Content: sysPrompt},
		{Role: "user", Content: question},
	}
	return CallNVIDIAChat(messages, "")
}

// GetNVIDIAModelList returns the built-in selectable model list as JSON string
func GetNVIDIAModelList() string {
	models := DefaultModels()
	data, _ := json.Marshal(models)
	return string(data)
}

// TestNVIDIAConnection tests connectivity and key validity; returns JSON result
func TestNVIDIAConnection(apiKey string, model string) string {
	if apiKey == "" {
		return `{"success":false,"error":"API Key 不能为空"}`
	}
	if model == "" {
		model = "z-ai/glm-5.1"
	}

	reqBody := NVIDIAChatRequest{
		Model:       model,
		Messages:    []NVIDIAChatMessage{{Role: "user", Content: "Hi"}},
		MaxTokens:   10,
		Temperature: 0,
	}

	jsonData, _ := json.Marshal(reqBody)
	url := NVIDIA_API_BASE + "/chat/completions"
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Sprintf(`{"success":false,"error":"连接失败: %v"}`, err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		bStr := string(body)
		if len(bStr) > 200 {
			bStr = bStr[:200]
		}
		return fmt.Sprintf(`{"success":false,"error":"HTTP %d: %s"}`, resp.StatusCode, bStr)
	}
	return `{"success":true,"message":"连接成功，API Key 有效","model":"` + model + `"}`
}
