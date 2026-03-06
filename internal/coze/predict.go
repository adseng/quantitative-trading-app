package coze

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/coze-dev/coze-go"
	"quantitative-trading-app/internal/binance"
	"quantitative-trading-app/internal/config"
	"quantitative-trading-app/internal/factor"
)

// cozeHTTPClient 返回禁用 Keep-Alive 的 HTTP 客户端，避免代理/服务器在空闲连接上返回 400 导致 "Unsolicited response on idle HTTP channel"
func cozeHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 90 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}
}

const (
	KlineCount    = 20
	KlineInterval = "15m"
)

// PredictResult 预测结果
type PredictResult struct {
	Answer   string
	TokenCnt int
}

// CozeStructuredResult 结构化预测结果
type CozeStructuredResult struct {
	Timestamp       string         `json:"timestamp"`
	Symbol          string         `json:"symbol"`
	CurrentPrice    float64        `json:"current_price"`
	MarketStructure string         `json:"market_structure"`
	Scenarios       []CozeScenario `json:"scenarios"`
	RawAnswer       string         `json:"-"` // 解析失败时保留原始回复
}

// CozeScenario 单个场景
type CozeScenario struct {
	Direction        string   `json:"direction"`
	Probability      int      `json:"probability"`
	SetupLogic       string   `json:"setup_logic"`
	TriggerCondition string   `json:"trigger_condition"`
	EntryPrice       *float64 `json:"entry_price"`
	StopLoss         *float64 `json:"stop_loss"`
	TakeProfit1      *float64 `json:"take_profit_1"`
	TakeProfit2      *float64 `json:"take_profit_2"`
	RiskRewardRatio  *float64 `json:"risk_reward_ratio"`
	Action           string   `json:"action"`
}

func formatKlinesForPrompt(klines []*factor.KLine) string {
	if len(klines) == 0 {
		return "[]"
	}
	type row struct {
		Idx   int     `json:"idx"`
		Time  int64   `json:"openTime"`
		Open  float64 `json:"open"`
		High  float64 `json:"high"`
		Low   float64 `json:"low"`
		Close float64 `json:"close"`
		Vol   float64 `json:"volume"`
	}
	rows := make([]row, len(klines))
	for i, k := range klines {
		rows[i] = row{
			Idx: i + 1, Time: k.OpenTime,
			Open: k.Open, High: k.High, Low: k.Low, Close: k.Close, Vol: k.Volume,
		}
	}
	b, _ := json.MarshalIndent(rows, "", "  ")
	return string(b)
}

// Predict 获取最近 20 根 15 分钟 K 线，调用 Coze 智能体预测
func Predict(ctx context.Context, symbol string) (*PredictResult, error) {
	if symbol == "" {
		symbol = config.Get(config.KeySymbol, "BTCUSDT")
	}
	binance.InitClient()
	klines, err := binance.FetchKlines(symbol, KlineInterval, KlineCount, nil)
	if err != nil {
		return nil, fmt.Errorf("获取 K 线失败: %w", err)
	}
	if len(klines) == 0 {
		return nil, fmt.Errorf("未获取到 K 线数据")
	}
	token := config.Get(config.KeyCozeAPIToken, "")
	botID := config.Get(config.KeyCozeBotID, "")
	if token == "" || botID == "" {
		return nil, fmt.Errorf("请在 .env.local 中配置 COZE_API_TOKEN 和 COZE_BOT_ID")
	}
	// 智能体侧提示词已配置，这里仅传递 K 线数据（D1）
	prompt := formatKlinesForPrompt(klines)

	authCli := coze.NewTokenAuth(token)
	baseURL := config.Get(config.KeyCozeBaseURL, "")
	opts := []coze.CozeAPIOption{coze.WithHttpClient(cozeHTTPClient())}
	if baseURL != "" {
		opts = append(opts, coze.WithBaseURL(baseURL))
	}
	cozeCli := coze.NewCozeAPI(authCli, opts...)
	req := &coze.CreateChatsReq{
		BotID:  botID,
		UserID: "cozepredict-user",
		Messages: []*coze.Message{
			coze.BuildUserQuestionText(prompt, nil),
		},
	}
	timeout := 60
	poll, err := cozeCli.Chat.CreateAndPoll(ctx, req, &timeout)
	if err != nil {
		return nil, fmt.Errorf("调用 Coze 智能体失败: %w", err)
	}
	var answer string
	for _, m := range poll.Messages {
		if m != nil && m.Role == coze.MessageRoleAssistant && m.Type == coze.MessageTypeAnswer {
			answer = strings.TrimSpace(m.Content)
			break
		}
	}
	if answer == "" {
		answer = "(无文本回复)"
	}
	tokenCnt := 0
	if poll.Chat != nil && poll.Chat.Usage != nil {
		tokenCnt = poll.Chat.Usage.TokenCount
	}
	return &PredictResult{Answer: answer, TokenCnt: tokenCnt}, nil
}

// PredictStructured 使用 K 线数据调用 Coze，请求结构化 JSON 预测结果。count 为发给豆包的 K 线根数，≤0 时默认 50。
func PredictStructured(ctx context.Context, klines []*factor.KLine, symbol string, count int) (*CozeStructuredResult, error) {
	if len(klines) == 0 {
		return nil, fmt.Errorf("K 线数据为空")
	}
	if symbol == "" {
		symbol = config.Get(config.KeySymbol, "BTCUSDT")
	}
	if count <= 0 {
		count = 50
	}
	n := count
	if len(klines) < n {
		n = len(klines)
	}
	recent := klines[len(klines)-n:]

	token := config.Get(config.KeyCozeAPIToken, "")
	botID := config.Get(config.KeyCozeBotID, "")
	if token == "" || botID == "" {
		return nil, fmt.Errorf("请在 .env.local 中配置 COZE_API_TOKEN 和 COZE_BOT_ID")
	}
	// 智能体侧提示词/工作流已配置，这里仅传递 K 线数据（D1）
	// 约定：智能体返回结构化 JSON（timestamp/symbol/current_price/market_structure/scenarios）
	prompt := formatKlinesForPrompt(recent)

	authCli := coze.NewTokenAuth(token)
	baseURL := config.Get(config.KeyCozeBaseURL, "")
	opts := []coze.CozeAPIOption{coze.WithHttpClient(cozeHTTPClient())}
	if baseURL != "" {
		opts = append(opts, coze.WithBaseURL(baseURL))
	}
	cozeCli := coze.NewCozeAPI(authCli, opts...)
	req := &coze.CreateChatsReq{
		BotID:  botID,
		UserID: "cozepredict-user",
		Messages: []*coze.Message{
			coze.BuildUserQuestionText(prompt+"\n请根据以上K线数据直接输出结构化预测JSON（仅输出JSON，不要其他内容）。", nil),
		},
	}
	timeout := 60
	poll, err := cozeCli.Chat.CreateAndPoll(ctx, req, &timeout)
	if err != nil {
		return nil, fmt.Errorf("调用 Coze 智能体失败: %w", err)
	}
	var answer string
	for _, m := range poll.Messages {
		if m != nil && m.Role == coze.MessageRoleAssistant && m.Type == coze.MessageTypeAnswer {
			answer = strings.TrimSpace(m.Content)
			break
		}
	}
	if answer == "" {
		return nil, fmt.Errorf("智能体无文本回复")
	}
	// 适配「仅输出 JSON」的 agent：去掉 markdown 代码块包裹，再取首尾 {} 之间的内容
	jsonStr := strings.TrimSpace(answer)
	if strings.HasPrefix(jsonStr, "```") {
		if first := strings.Index(jsonStr, "\n"); first >= 0 {
			jsonStr = strings.TrimSpace(jsonStr[first+1:])
		}
		if end := strings.Index(jsonStr, "```"); end >= 0 {
			jsonStr = strings.TrimSpace(jsonStr[:end])
		}
	}
	if idx := strings.Index(jsonStr, "{"); idx >= 0 {
		if end := strings.LastIndex(jsonStr, "}"); end > idx {
			jsonStr = jsonStr[idx : end+1]
		}
	}
	var out CozeStructuredResult
	if err := json.Unmarshal([]byte(jsonStr), &out); err != nil {
		return &CozeStructuredResult{RawAnswer: answer}, nil
	}
	return &out, nil
}

// recordExcelSafe 异步写入 Excel，内部 recover 防止 panic 影响主流程
func recordExcelSafe(res *CozeStructuredResult) {
	defer func() { _ = recover() }()
	_ = AppendResultToExcel(res)
}

// PredictStructuredWithNotify 调用 PredictStructured，按顺序回调 onStatus（requesting/error/done）与 onResult，并异步写入 Excel；带 recover，回调可为 nil。供 app 层复用。
func PredictStructuredWithNotify(
	ctx context.Context,
	klines []*factor.KLine,
	symbol string,
	count int,
	onStatus func(status, message string),
	onResult func(*CozeStructuredResult),
) (res *CozeStructuredResult, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
			if onStatus != nil {
				onStatus("error", err.Error())
			}
		}
	}()
	if onStatus != nil {
		onStatus("requesting", "")
	}
	res, err = PredictStructured(ctx, klines, symbol, count)
	if err != nil {
		if onStatus != nil {
			onStatus("error", err.Error())
		}
		return nil, err
	}
	if onStatus != nil {
		onStatus("done", "")
	}
	if onResult != nil {
		onResult(res)
	}
	go recordExcelSafe(res)
	return res, nil
}
