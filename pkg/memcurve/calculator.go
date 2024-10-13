package memcurve

import (
	"cmp"
	"time"
)

// DefaultIntervals 默认复习间隔时间，可以作为 NewReviewCalculator 函数的参数，或者可以直接使用
var DefaultIntervals = []time.Duration{
	20 * time.Minute,    // 20 分钟后
	40 * time.Minute,    // 1 小时后
	540 * time.Minute,   // 9 小时后
	24 * time.Hour,      // 1 天后
	2 * 24 * time.Hour,  // 2 天后
	4 * 24 * time.Hour,  // 4 天后
	7 * 24 * time.Hour,  // 7 天后
	14 * 24 * time.Hour, // 2 周后
	28 * 24 * time.Hour, // 长期复习状态，每月过一遍
}

const (
	ConfidenceCertain = 1   // 牢记
	ConfidenceHigh    = 0.8 // 清晰 学习者非常确定记得知识
	ConfidenceMedium  = 0.5 // 熟悉 学习者有一定的把握记得知识
	ConfidenceLow     = 0.3 // 模糊 学习者对知识点的记忆不太确定
	ConfidenceNone    = 0.1 // 陌生 学习者几乎完全遗忘了知识

	// ReasonableConfidence 定义了合理的置信度阈值，当用户置信度高于该值时，表示记忆状态良好
	ReasonableConfidence = 0.6
)

// ReviewCalculator 结构体它集成了复习时间的计算，考虑了置信度和用户参数
type ReviewCalculator struct {
	reviewIntervals []time.Duration // 复习间隔
	userFactors     UserFactors     // 用户个性化的因素
	ReviewLevel     int             // 每个知识点的当前复习级别
}

// UserFactors 表示影响个体遗忘速率的参数
type UserFactors struct {
	// ForgettingSpeed 表示遗忘速度
	// 1 表示正常速度，不会增加或减少 interval
	// 最少是 0.2 (表示遗忘的很慢，可以有 5 倍的 interval)
	// 最多是 5 (表示遗忘的很快)
	ForgettingSpeed float64
}

// ReviewData 包含复习记录和用户因素
type ReviewData struct {
	ReviewRecords []time.Time // 复习记录
	ReviewLevel   int         // 每个知识点的当前复习级别
}

// NewReviewCalculator 创建一个新的ReviewCalculator实例
func NewReviewCalculator(intervals []time.Duration, factors UserFactors) *ReviewCalculator {
	if len(intervals) == 0 {
		intervals = DefaultIntervals
	}
	return &ReviewCalculator{
		reviewIntervals: intervals,
		userFactors:     factors,
	}
}

// CalculateNextReview 根据上次复习时间、复习次数和置信度计算下一次复习的时间
// 现在返回新的复习级别并适配调整间隔的逻辑
func (calc *ReviewCalculator) CalculateNextReview(data *ReviewData, newConfidence float64) (newReviewTime time.Time, newReviewLevel int) {
	lastReview := time.Now()
	if len(data.ReviewRecords) > 0 {
		lastReview = data.ReviewRecords[len(data.ReviewRecords)-1]
	}

	// 计算复习级别
	newReviewLevel = calc.CalculateReviewLevel(data, newConfidence)

	// 根据复习级别，计算当前应复习的间隔时间
	baseInterval := calc.ReviewIntervalOf(newReviewLevel)

	// 获取上一次复习间隔，用于结合置信度修正遗忘速度
	var lastInterval time.Duration
	if len(data.ReviewRecords) >= 2 {
		lastInterval = lastReview.Sub(data.ReviewRecords[len(data.ReviewRecords)-2])
	} else {
		lastInterval = 0
	}

	// 使用新的置信度调整函数
	adjustedInterval := calc.adjustIntervalWithConfidence(baseInterval, lastInterval, newConfidence)
	newReviewTime = lastReview.Add(adjustedInterval)

	return newReviewTime, newReviewLevel
}

// CalculateReviewLevel 根据上次一复习级别和置信度，计算下次的合理复习级别，
func (calc *ReviewCalculator) CalculateReviewLevel(data *ReviewData, newConfidence float64) int {
	// 如果置信度低，可能意味着用户需要重新开始复习过程，返回到较低的复习级别
	// 如果置信度高，用户可能已经巩固了知识，可以提升复习级别

	currentLevel := calc.clampReviewLevel(data.ReviewLevel)
	if newConfidence <= ConfidenceNone { // 低置信度重置复习级别到 0
		return 0
	} else if newConfidence <= ConfidenceLow { // 低置信度降低复习级别
		return calc.clampReviewLevel(max(1, currentLevel/2))
	} else if newConfidence <= ConfidenceMedium {
		// 中置信度少量推进复习级别，但不会推进到最后2级
		// todo: provide config
		return calc.clampReviewLevel(min(len(calc.reviewIntervals)-3, currentLevel+1))
	} else if newConfidence <= ConfidenceHigh {
		// 高置信度，按部就班的增加复习级别，但不应超过当前记录的复习次数
		return calc.clampReviewLevel(min(len(calc.reviewIntervals)-1, currentLevel+1))
	} else {
		// 牢记级别，直接 x 2，但不应超过当前记录的复习次数
		return calc.clampReviewLevel(min(len(calc.reviewIntervals)-1, max(currentLevel, 1)*2))
	}
}

// MaxReviewLevel 最高的 Review 级别，-1 表示无配置
func (calc *ReviewCalculator) MaxReviewLevel() int {
	return len(calc.reviewIntervals) - 1
}

// ReviewIntervalOf 获取复习间隔，防止索引超出范围
func (calc *ReviewCalculator) ReviewIntervalOf(level int) time.Duration {
	return calc.reviewIntervals[calc.clampReviewLevel(level)]
}

func (calc *ReviewCalculator) clampReviewLevel(level int) int {
	return clamp(level, 0, calc.MaxReviewLevel())
}

// adjustoFrgettingSpeedWithConfidence 根据复习间隔和置信度来推测遗忘速度，以调整复习时间表
func (calc *ReviewCalculator) adjustIntervalWithConfidence(
	setInterval time.Duration, realInterval time.Duration,
	confidence float64,
) time.Duration {
	// 防止除以 0 的情况，使用初始遗忘值
	if calc.userFactors.ForgettingSpeed == 0 {
		calc.userFactors.ForgettingSpeed = 1
	}

	calc.userFactors.ForgettingSpeed = adjustoFrgettingSpeedWithConfidence(
		calc.userFactors.ForgettingSpeed,
		setInterval, realInterval, confidence,
	)

	return calcIntervalByForgettingSpeed(setInterval, calc.userFactors.ForgettingSpeed)
}

// calcIntervalByForgettingSpeed 根据推测的遗忘速度调整复习间隔
// todo: 优化算法，当前是简单倒数
func calcIntervalByForgettingSpeed(duration time.Duration, forgettingSpeed float64) time.Duration {
	if forgettingSpeed == 0 {
		forgettingSpeed = 1
	}
	return time.Duration(float64(duration) / forgettingSpeed)
}

func adjustoFrgettingSpeedWithConfidence(
	forgettingSpeed float64,
	setInterval time.Duration, realInterval time.Duration,
	confidence float64,
) float64 {
	forgettingSpeed = clamp(forgettingSpeed, 0.2, 5.0)
	if realInterval <= 0 {
		return forgettingSpeed
	}

	// 换算当前遗忘速度下的设定间隔
	expectedInterval := float64(calcIntervalByForgettingSpeed(setInterval, forgettingSpeed))

	// 置信度与合理置信度的比较来推测遗忘速度
	// todo: 现在的 ReasonableConfidence 二值判断改成模态
	if confidence > ReasonableConfidence && float64(realInterval) > expectedInterval {
		// 置信度高于合理的置信度, 但用户复习间隔比预计的更长。说明过了更长的时间仍然记得，遗忘速度可能被高估了
		forgettingSpeed *= expectedInterval / float64(realInterval)
	} else if confidence < ReasonableConfidence && float64(realInterval) < expectedInterval {
		// 用户复习间隔比预计的短，却忘记了，可能意味着遗忘速度被低估了
		forgettingSpeed *= expectedInterval / float64(realInterval)
	}

	// 再次确保遗忘速度在有意义的范围内
	return clamp(forgettingSpeed, 0.2, 5.0)
}

// clamp 函数用于确保遗忘速度在合理的范围内
func clamp[T cmp.Ordered](value, minV, maxV T) T {
	return max(minV, min(maxV, value))
}
