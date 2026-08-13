package main

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/XIU2/CloudflareSpeedTest/utils"
)

const (
	defaultUploadConfigFile = "cfyx.json"
	defaultGitHubFilePath   = "cfyx.txt"
	githubAPIBaseURL        = "https://api.github.com"
	maxAPIResponseSize      = 2 * 1024 * 1024
	maxResultRows           = 10000
)

type uploadConfig struct {
	CFAPIToken     string `json:"cf_api_token"`
	CFZoneID       string `json:"cf_zone_id"`
	CFBaseDomain   string `json:"cf_base_domain"`
	CFProxied      bool   `json:"cf_proxied"`
	GitHubToken    string `json:"github_token"`
	GitHubRepo     string `json:"github_repo"`
	GitHubFilePath string `json:"github_file_path"`
}

func defaultUploadConfig() uploadConfig {
	return uploadConfig{
		CFProxied:      true,
		GitHubFilePath: defaultGitHubFilePath,
	}
}

func loadUploadConfig() (uploadConfig, error) {
	data, err := os.ReadFile(defaultUploadConfigFile)
	if os.IsNotExist(err) {
		cfg := defaultUploadConfig()
		_ = saveUploadConfig(cfg)
		return cfg, fmt.Errorf("配置文件不存在，已创建 %s，请填写后重试", defaultUploadConfigFile)
	}
	if err != nil {
		return uploadConfig{}, fmt.Errorf("读取配置文件失败: %v", err)
	}
	var cfg uploadConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return uploadConfig{}, fmt.Errorf("解析配置文件失败: %v", err)
	}
	if cfg.GitHubFilePath == "" {
		cfg.GitHubFilePath = defaultGitHubFilePath
	}
	return cfg, nil
}

func saveUploadConfig(cfg uploadConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(defaultUploadConfigFile, data, 0600)
}

func ensureUploadConfigFile() {
	_, err := os.Stat(defaultUploadConfigFile)
	if err == nil {
		return
	}
	if !os.IsNotExist(err) {
		utils.Yellow.Printf("检查 %s 失败: %v\n", defaultUploadConfigFile, err)
		return
	}

	cfg := defaultUploadConfig()
	if err := saveUploadConfig(cfg); err != nil {
		utils.Yellow.Printf("创建默认配置文件 %s 失败: %v\n", defaultUploadConfigFile, err)
		return
	}
	utils.Green.Printf("已创建默认配置文件 %s，请填写 API 配置后重新运行。\n", defaultUploadConfigFile)
}

func maybeUpload(speedData utils.DownloadSpeedSet) {
	if len(speedData) == 0 {
		return
	}

	fmt.Println()
	fmt.Println("是否上传优选 IP？")
	fmt.Println("1. 不上传（默认）")
	fmt.Println("2. Cloudflare DNS")
	fmt.Println("3. GitHub 文件")
	fmt.Print("请选择：")

	mode := strings.ToLower(readPromptLine())
	switch mode {
	case "2", "cloudflare":
		ips := selectUploadIPs(speedData)
		if len(ips) == 0 {
			return
		}
		uploadCloudflareDNS(ips)
	case "3", "github":
		ips := selectUploadIPs(speedData)
		if len(ips) == 0 {
			return
		}
		uploadGitHubFile(ips)
	default:
		return
	}
}

func uploadResultFile() {
	fp, err := os.Open("result.csv")
	if os.IsNotExist(err) {
		utils.Yellow.Println("未找到 result.csv，请先执行测速。")
		return
	}
	if err != nil {
		utils.Red.Printf("读取 result.csv 失败: %v\n", err)
		return
	}
	defer fp.Close()

	reader := csv.NewReader(fp)
	if _, err := reader.Read(); err != nil {
		utils.Red.Printf("解析 result.csv 失败: %v\n", err)
		return
	}
	var speedData utils.DownloadSpeedSet
	for rowCount := 0; rowCount < maxResultRows; rowCount++ {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			utils.Red.Printf("解析 result.csv 失败: %v\n", err)
			return
		}
		if len(record) == 0 {
			continue
		}
		ip := net.ParseIP(strings.TrimSpace(record[0]))
		if ip == nil {
			continue
		}
		speedData = append(speedData, utils.CloudflareIPData{
			PingData: &utils.PingData{IP: &net.IPAddr{IP: ip}},
		})
	}

	if len(speedData) == 0 {
		utils.Yellow.Println("result.csv 中没有有效优选 IP，请先执行测速。")
		return
	}
	if _, err := reader.Read(); err == nil {
		utils.Yellow.Printf("result.csv 仅读取前 %d 行数据。\n", maxResultRows)
	} else if err != io.EOF {
		utils.Red.Printf("解析 result.csv 失败: %v\n", err)
		return
	}
	printUploadCandidates(speedData)

	maybeUpload(speedData)
}

func printUploadCandidates(speedData utils.DownloadSpeedSet) {
	count := utils.PrintNum
	if count > len(speedData) {
		count = len(speedData)
	}
	if count <= 0 {
		return
	}

	fmt.Println("\n已有优选 IP：")
	for i := 0; i < count; i++ {
		fmt.Printf("%d. %s\n", i+1, speedData[i].IP)
	}
}

func selectUploadIPs(speedData utils.DownloadSpeedSet) []string {
	maxSelectable := utils.PrintNum
	if maxSelectable <= 0 {
		fmt.Println("未显示测速结果，跳过上传。")
		return nil
	}
	if maxSelectable > len(speedData) {
		maxSelectable = len(speedData)
	}

	fmt.Println()
	fmt.Print("请选择要上传的优选 IP 编号，多个用英文逗号分隔，按 ENTER 跳过：")

	raw := readPromptLine()
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	var selected []string
	for _, part := range strings.Split(raw, ",") {
		idx, err := strconv.Atoi(strings.TrimSpace(part))
		if err == nil && idx >= 1 && idx <= maxSelectable {
			selected = append(selected, speedData[idx-1].IP.String())
		}
	}

	if len(selected) == 0 {
		fmt.Println("未选择有效 IP，跳过上传。")
	}
	return selected
}

func uploadCloudflareDNS(ips []string) {
	cfg, err := loadUploadConfig()
	if err != nil {
		utils.Yellow.Println(err)
		return
	}
	if cfg.CFAPIToken == "" || cfg.CFZoneID == "" || cfg.CFBaseDomain == "" {
		utils.Red.Println("Cloudflare 上传配置不完整，请检查 cfyx.json 中的 cf_api_token、cf_zone_id、cf_base_domain。")
		return
	}

	baseDomain := strings.TrimPrefix(strings.TrimSpace(cfg.CFBaseDomain), ".")
	for i, ip := range ips {
		domain := fmt.Sprintf("yx%d.%s", i+1, baseDomain)
		if err := upsertCloudflareDNS(cfg, domain, ip); err != nil {
			utils.Red.Printf("Cloudflare %s 更新失败: %v\n", domain, err)
			continue
		}
		utils.Green.Printf("Cloudflare %s -> %s 已更新\n", domain, ip)
	}
}

type cfDNSBody struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
	Proxied bool   `json:"proxied"`
}

type cfListResponse struct {
	Success bool `json:"success"`
	Errors  []struct {
		Message string `json:"message"`
	} `json:"errors"`
	Result []struct {
		ID string `json:"id"`
	} `json:"result"`
}

func recordTypeForIP(ip string) string {
	if strings.Contains(ip, ":") {
		return "AAAA"
	}
	return "A"
}

func cloudflareRequest(cfg uploadConfig, method, apiURL string, body []byte) ([]byte, int, error) {
	req, err := http.NewRequest(method, apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.CFAPIToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	data, err := readLimitedResponse(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return data, resp.StatusCode, nil
}

func readLimitedResponse(body io.Reader) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(body, maxAPIResponseSize+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxAPIResponseSize {
		return nil, fmt.Errorf("响应内容超过 %d MB 限制", maxAPIResponseSize/(1024*1024))
	}
	return data, nil
}

func upsertCloudflareDNS(cfg uploadConfig, domain, ip string) error {
	recordType := recordTypeForIP(ip)
	baseURL := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records", url.PathEscape(cfg.CFZoneID))

	queryURL := fmt.Sprintf("%s?name=%s&type=%s", baseURL, url.QueryEscape(domain), url.QueryEscape(recordType))
	body, status, err := cloudflareRequest(cfg, http.MethodGet, queryURL, nil)
	if err != nil {
		return err
	}

	var listResp cfListResponse
	if err := json.Unmarshal(body, &listResp); err != nil {
		return fmt.Errorf("解析 Cloudflare 查询结果失败: %v", err)
	}
	if status != 200 || !listResp.Success {
		if len(listResp.Errors) > 0 {
			return fmt.Errorf("Cloudflare 查询失败: %s", listResp.Errors[0].Message)
		}
		return fmt.Errorf("Cloudflare 查询失败，HTTP %d", status)
	}

	payload, err := json.Marshal(cfDNSBody{
		Type:    recordType,
		Name:    domain,
		Content: ip,
		TTL:     1,
		Proxied: cfg.CFProxied,
	})
	if err != nil {
		return err
	}

	if len(listResp.Result) > 0 {
		recordID := listResp.Result[0].ID
		updateURL := baseURL + "/" + url.PathEscape(recordID)
		_, status, err := cloudflareRequest(cfg, http.MethodPatch, updateURL, payload)
		if err != nil {
			return err
		}
		if status != 200 {
			return fmt.Errorf("Cloudflare 更新失败，HTTP %d", status)
		}
		return nil
	}

	_, status, err = cloudflareRequest(cfg, http.MethodPost, baseURL, payload)
	if err != nil {
		return err
	}
	if status != 200 {
		return fmt.Errorf("Cloudflare 创建失败，HTTP %d", status)
	}
	return nil
}

type gitHubContentResponse struct {
	Content string `json:"content"`
	Sha     string `json:"sha"`
}

func uploadGitHubFile(ips []string) {
	cfg, err := loadUploadConfig()
	if err != nil {
		utils.Yellow.Println(err)
		return
	}
	if cfg.GitHubToken == "" || cfg.GitHubRepo == "" {
		utils.Red.Println("GitHub 上传配置不完整，请检查 cfyx.json 中的 github_token、github_repo。")
		return
	}
	if err := checkGitHubConnectivity(); err != nil {
		utils.Red.Printf("GitHub 网络连接失败: %v\n", err)
		printGitHubNetworkHint()
		return
	}

	path := strings.TrimSpace(cfg.GitHubFilePath)
	if path == "" {
		path = defaultGitHubFilePath
	}

	oldContent, sha, exists, err := getGitHubFile(cfg, path)
	if err != nil {
		utils.Red.Printf("读取 GitHub 文件失败: %v\n", err)
		if isNetworkError(err) {
			printGitHubNetworkHint()
		}
		return
	}

	newContent := mergeGitHubIPContent(oldContent, ips)
	contentBase64 := base64.StdEncoding.EncodeToString([]byte(newContent))

	body := map[string]string{
		"message": "update optimized ip",
		"content": contentBase64,
	}
	if exists && sha != "" {
		body["sha"] = sha
	}
	payload, err := json.Marshal(body)
	if err != nil {
		utils.Red.Printf("GitHub 上传请求构造失败: %v\n", err)
		return
	}

	apiURL := fmt.Sprintf("%s/repos/%s/contents/%s", githubAPIBaseURL, cfg.GitHubRepo, escapeGitHubPath(path))
	req, err := http.NewRequest(http.MethodPut, apiURL, bytes.NewReader(payload))
	if err != nil {
		utils.Red.Printf("GitHub 上传请求失败: %v\n", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+cfg.GitHubToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		utils.Red.Printf("GitHub 上传失败: %v\n", err)
		if isNetworkError(err) {
			printGitHubNetworkHint()
		}
		return
	}
	defer resp.Body.Close()
	respBody, readErr := readLimitedResponse(resp.Body)
	if readErr != nil {
		utils.Red.Printf("读取 GitHub 响应失败: %v\n", readErr)
		return
	}
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		utils.Red.Printf("GitHub 上传失败，HTTP %d: %s\n", resp.StatusCode, strings.TrimSpace(string(respBody)))
		return
	}

	utils.Green.Printf("GitHub 文件 %s 已更新，包含 %d 个优选 IP\n", path, len(ips))
}

func checkGitHubConnectivity() error {
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get(githubAPIBaseURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func isNetworkError(err error) bool {
	var netErr net.Error
	return err != nil && errors.As(err, &netErr)
}

func printGitHubNetworkHint() {
	utils.Yellow.Println("请检查网络连接；在无法直连 GitHub 的网络环境中，请先开启本地代理后重试。")
}

func getGitHubFile(cfg uploadConfig, path string) (string, string, bool, error) {
	apiURL := fmt.Sprintf("%s/repos/%s/contents/%s", githubAPIBaseURL, cfg.GitHubRepo, escapeGitHubPath(path))
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return "", "", false, err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.GitHubToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", false, err
	}
	defer resp.Body.Close()
	body, err := readLimitedResponse(resp.Body)
	if err != nil {
		return "", "", false, fmt.Errorf("读取 GitHub 响应失败: %v", err)
	}
	if resp.StatusCode == 404 {
		return "", "", false, nil
	}
	if resp.StatusCode != 200 {
		return "", "", false, fmt.Errorf("GitHub GET 失败，HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var contentResp gitHubContentResponse
	if err := json.Unmarshal(body, &contentResp); err != nil {
		return "", "", false, err
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(contentResp.Content, "\n", ""))
	if err != nil {
		return "", "", false, err
	}
	return string(decoded), contentResp.Sha, true, nil
}

func escapeGitHubPath(path string) string {
	escaped := url.PathEscape(path)
	return strings.ReplaceAll(escaped, "%2F", "/")
}

func mergeGitHubIPContent(oldContent string, ips []string) string {
	normalized := strings.ReplaceAll(oldContent, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")

	var out []string
	ipIndex := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if net.ParseIP(trimmed) != nil {
			if ipIndex < len(ips) {
				out = append(out, ips[ipIndex])
				ipIndex++
			}
			continue
		}
		out = append(out, line)
	}

	for ipIndex < len(ips) {
		out = append(out, ips[ipIndex])
		ipIndex++
	}

	return strings.Join(out, "\n") + "\n"
}
