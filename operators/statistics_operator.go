package operators

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/zhanghuachuan/water-reminder/database"
	"github.com/zhanghuachuan/water-reminder/utils"
)

type StatisticsOperator struct{}

func (o *StatisticsOperator) Name() string {
	return "statistics"
}

type StatisticsRequest struct {
	Period string `json:"period"` // day/week/month/custom
	Date   string `json:"date"`   // 基准日期，格式 YYYY-MM-DD
	Start  string `json:"start"`  // 自定义开始日期（当period=custom时使用）
	End    string `json:"end"`    // 自定义结束日期（当period=custom时使用）
}

type StatisticsResponse struct {
	Period           string            `json:"period"`           // 统计周期
	StartDate        string            `json:"startDate"`        // 统计开始日期
	EndDate          string            `json:"endDate"`          // 统计结束日期
	TotalAmount      float64           `json:"totalAmount"`      // 总饮水量（毫升）
	DailyAverage     float64           `json:"dailyAverage"`     // 日均饮水量
	DailyGoal        float64           `json:"dailyGoal"`        // 每日目标
	Progress         float64           `json:"progress"`         // 完成百分比
	DrinkTypes       map[string]int    `json:"drinkTypes"`       // 饮品类型分布
	TimeDistribution map[string]int    `json:"timeDistribution"` // 时间段分布（上午/下午/晚上）
	HourlySummary    map[string]int    `json:"hourlySummary"`    // 按小时统计
	Records          []WaterRecordInfo `json:"records"`          // 详细记录
	Message          string            `json:"message"`          // 提示信息
}

type WaterRecordInfo struct {
	Time      time.Time `json:"time"`
	Amount    float64   `json:"amount"`
	DrinkType string    `json:"drinkType"`
}

func (o *StatisticsOperator) Execute(ctx context.Context, w http.ResponseWriter, r *http.Request) (context.Context, error) {
	// 获取当前用户
	user, ok := ctx.Value("user").(*utils.User)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return ctx, nil
	}

	var req StatisticsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return ctx, err
	}

	// 验证请求参数
	if req.Period == "" {
		req.Period = "day"
	}
	if !utils.Contains([]string{"day", "week", "month", "custom"}, req.Period) {
		http.Error(w, "Invalid period. Allowed values: day, week, month, custom", http.StatusBadRequest)
		return ctx, nil
	}

	if req.Date == "" {
		req.Date = time.Now().Format("2006-01-02")
	}
	if _, err := time.Parse("2006-01-02", req.Date); err != nil {
		http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return ctx, nil
	}

	if req.Period == "custom" {
		if req.Start == "" || req.End == "" {
			http.Error(w, "Start and end dates are required for custom period", http.StatusBadRequest)
			return ctx, nil
		}
		if _, err := time.Parse("2006-01-02", req.Start); err != nil {
			http.Error(w, "Invalid start date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return ctx, nil
		}
		if _, err := time.Parse("2006-01-02", req.End); err != nil {
			http.Error(w, "Invalid end date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return ctx, nil
		}
	}

	// 处理统计数据
	response := o.generateStatistics(req, user)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	return ctx, nil
}

func (o *StatisticsOperator) generateStatistics(req StatisticsRequest, user *utils.User) StatisticsResponse {
	// 解析日期范围
	baseDate, _ := time.Parse("2006-01-02", req.Date)
	startTime := baseDate
	endTime := baseDate.Add(24 * time.Hour)

	switch req.Period {
	case "week":
		startTime = baseDate.AddDate(0, 0, -6)
	case "month":
		startTime = baseDate.AddDate(0, -1, 0)
	case "custom":
		startTime, _ = time.Parse("2006-01-02", req.Start)
		endTime, _ = time.Parse("2006-01-02", req.End)
		endTime = endTime.Add(24 * time.Hour)
	}

	// 从数据库查询记录
	var records []struct {
		RecordTime time.Time
		Amount     float64
		DrinkType  string
	}
	db := database.GetDB()
	db.Table("water_records").
		Select("record_time, amount, drink_type").
		Where("user_id = ? AND record_time BETWEEN ? AND ?",
			user.ID, startTime, endTime).
		Scan(&records)

	// 转换为WaterRecordInfo格式
	var recordInfos []WaterRecordInfo
	for _, r := range records {
		recordInfos = append(recordInfos, WaterRecordInfo{
			Time:      r.RecordTime,
			Amount:    r.Amount,
			DrinkType: r.DrinkType,
		})
	}

	// 计算统计数据
	totalAmount := 0.0
	drinkTypes := make(map[string]int)
	hourlySummary := make(map[string]int)

	for _, record := range recordInfos {
		totalAmount += record.Amount
		drinkTypes[record.DrinkType]++
		hour := record.Time.Format("15:00")
		hourlySummary[hour] += int(record.Amount)
	}

	// 计算时间段分布
	timeDistribution := map[string]int{
		"morning":   0, // 6-12
		"afternoon": 0, // 12-18
		"evening":   0, // 18-24
		"night":     0, // 0-6
	}

	for _, record := range recordInfos {
		hour := record.Time.Hour()
		switch {
		case hour >= 6 && hour < 12:
			timeDistribution["morning"] += int(record.Amount)
		case hour >= 12 && hour < 18:
			timeDistribution["afternoon"] += int(record.Amount)
		case hour >= 18 && hour < 24:
			timeDistribution["evening"] += int(record.Amount)
		default:
			timeDistribution["night"] += int(record.Amount)
		}
	}

	// 计算统计指标
	dailyGoal := 2000.0
	days := 1.0
	if req.Period == "week" {
		days = 7
	} else if req.Period == "month" {
		days = 30 // 简化处理，实际应按月份天数计算
	}

	progress := (totalAmount / (dailyGoal * days)) * 100
	if progress > 100 {
		progress = 100
	}

	// 设置响应日期范围
	startDate := req.Date
	endDate := req.Date

	if req.Period == "week" {
		startDate = baseDate.AddDate(0, 0, -6).Format("2006-01-02")
	} else if req.Period == "month" {
		startDate = baseDate.AddDate(0, -1, 0).Format("2006-01-02")
	} else if req.Period == "custom" {
		startDate = req.Start
		endDate = req.End
	}

	return StatisticsResponse{
		Period:           req.Period,
		StartDate:        startDate,
		EndDate:          endDate,
		TotalAmount:      totalAmount,
		DailyAverage:     totalAmount / days,
		DailyGoal:        dailyGoal,
		Progress:         progress,
		DrinkTypes:       drinkTypes,
		TimeDistribution: timeDistribution,
		HourlySummary:    hourlySummary,
		Records:          recordInfos,
		Message:          o.getMotivationMessage(progress),
	}
}

// 根据进度生成激励消息
func (o *StatisticsOperator) getMotivationMessage(progress float64) string {
	switch {
	case progress >= 100:
		return "恭喜！您已达成今日目标！"
	case progress >= 80:
		return "做得好！快完成目标了！"
	case progress >= 50:
		return "继续努力，您已经完成了一半！"
	default:
		return "记得多喝水哦！"
	}
}
