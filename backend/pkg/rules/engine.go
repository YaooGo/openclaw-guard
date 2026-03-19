package rules

import (
	"regexp"
	"strings"

	"github.com/openclaw-guard/backend/internal/config"
)

type Engine struct {
	config *config.Config

	// PII 检测正则
	phoneRegex       *regexp.Regexp
	idCardRegex      *regexp.Regexp
	apiKeyRegex      *regexp.Regexp
	ipv4Regex        *regexp.Regexp

	// 注入检测模式
	injectionPatterns []string

	// 危险操作模式
	dangerousPatterns []string
}

func NewEngine(cfg *config.Config) *Engine {
	return &Engine{
		config: cfg,
		phoneRegex: regexp.MustCompile(`\b1[3-9]\d{9}\b`),
		idCardRegex: regexp.MustCompile(`\b[1-9]\d{5}(18|19|20)\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])\d{3}[\dXx]\b`),
		apiKeyRegex: regexp.MustCompile(`\b(sk-|api_key|AKIA|ghp_|gho_|ghu_|ghs_|ghr_)[a-zA-Z0-9]{20,}\b`),
		ipv4Regex: regexp.MustCompile(`\b(?:(?:10|172\.(?:1[6-9]|2\d|3[01])|192\.168)\..*?\b|127\.0\.0\.1\b)`),

		injectionPatterns: []string{
			"ignore previous instructions",
			"ignore all instructions",
			"override system",
			"jailbreak",
			"dan mode",
			"developer mode",
			"above rules",
			"bypass security",
			"ignore security",
		},

		dangerousPatterns: []string{
			"rm -rf /",
			"rm -rf",
			"rm -r /",
			"mkfs",
			"dd if=/dev/zero",
			"dd if=/dev/random",
			":(){ :|:& };:",
			"chmod 000",
			"chown root:root",
			"format c:",
			"del /f /q",
			"drop database",
			"truncate table",
			"delete from",
		},
	}
}

// ScanInput 扫描输入内容，返回检测到的风险
func (e *Engine) ScanInput(input string) []string {
	risks := make([]string, 0)

	// PII 检测
	if e.phoneRegex.MatchString(input) {
		risks = append(risks, "包含手机号")
	}
	if e.idCardRegex.MatchString(input) {
		risks = append(risks, "包含身份证号")
	}
	if e.apiKeyRegex.MatchString(input) {
		risks = append(risks, "可能包含 API Key")
	}
	if e.ipv4Regex.MatchString(input) {
		risks = append(risks, "包含 IP 地址")
	}

	// 注入检测
	lowerInput := strings.ToLower(input)
	for _, pattern := range e.injectionPatterns {
		if strings.Contains(lowerInput, pattern) {
			risks = append(risks, "可能的注入攻击")
			break
		}
	}

	// 危险操作检测
	for _, pattern := range e.dangerousPatterns {
		if strings.Contains(lowerInput, pattern) {
			risks = append(risks, "危险操作指令")
			break
		}
	}

	return risks
}

// MaskPII 脱敏处理
func (e *Engine) MaskPII(input string) string {
	result := input

	// 脱敏手机号
	result = e.phoneRegex.ReplaceAllString(result, "138****5678")

	// 脱敏身份证
	result = e.idCardRegex.ReplaceAllString(result, "11010119900101****")

	// 脱敏 API Key
	result = e.apiKeyRegex.ReplaceAllString(result, "$1****")

	// 脱敏 IP
	result = e.ipv4Regex.ReplaceAllString(result, "xxx.xxx.xxx.xxx")

	return result
}

// CheckInjection 检查是否为注入攻击
func (e *Engine) CheckInjection(input string) bool {
	lowerInput := strings.ToLower(input)
	for _, pattern := range e.injectionPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}
	return false
}

// IsDangerousCommand 检查是否为危险命令
func (e *Engine) IsDangerousCommand(command string) bool {
	lowerCmd := strings.ToLower(command)
	for _, pattern := range e.dangerousPatterns {
		if strings.Contains(lowerCmd, pattern) {
			return true
		}
	}
	return false
}
