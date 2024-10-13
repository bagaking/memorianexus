package memcurve

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// 基础间隔生成测试
func TestCalculateNextReview(t *testing.T) {
	// 通过 NewReviewCalculator 创建一个默认的 reviewIntervals 和 userFactors 的实例
	calculator := NewReviewCalculator(nil, UserFactors{ForgettingSpeed: 1})
	now := time.Now()

	testCases := []struct {
		name                string
		currentReviewLevel  int
		inputConfidence     float64
		expectedNextLevel   int
		timeSinceLastReview time.Duration // 上次复习距现在的时间
	}{
		{
			name:                "Initial level with high confidence increases level",
			currentReviewLevel:  0,
			inputConfidence:     ConfidenceHigh,
			expectedNextLevel:   1,
			timeSinceLastReview: DefaultIntervals[0],
		},
		{
			name:                "Maximum level with high confidence stays same",
			currentReviewLevel:  calculator.MaxReviewLevel(),
			inputConfidence:     ConfidenceHigh,
			expectedNextLevel:   calculator.MaxReviewLevel(),
			timeSinceLastReview: DefaultIntervals[calculator.MaxReviewLevel()],
		},
		{
			name:                "Middle level with low confidence decreases level",
			currentReviewLevel:  3,
			inputConfidence:     ConfidenceLow,
			expectedNextLevel:   1,
			timeSinceLastReview: DefaultIntervals[3],
		},
		{
			name:                "Middle level with none confidence resets level",
			currentReviewLevel:  3,
			inputConfidence:     ConfidenceNone,
			expectedNextLevel:   0,
			timeSinceLastReview: DefaultIntervals[3],
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 创建 ReviewData 实例
			// 添加上一次复习的记录和当前复习级别
			data := ReviewData{
				ReviewRecords: []time.Time{now.Add(-tc.timeSinceLastReview)},
				ReviewLevel:   tc.currentReviewLevel,
			}

			// reset the calculator
			calculator = NewReviewCalculator(nil, UserFactors{ForgettingSpeed: 1})
			// 调用 CalculateNextReview 得到下次复习时间和新的复习级别
			nextReviewTime, nextReviewLevel := calculator.CalculateNextReview(&data, tc.inputConfidence)

			// 使用 testify 的 assert 函数来检查期望的复习级别
			assert.Equal(t, tc.expectedNextLevel, nextReviewLevel, "Expected next review level did not match")

			// 确认下次复习时间是否符合预期时长，为此我们获取应有的间隔并计算预期的下次复习时间
			expectedReviewInterval := calculator.ReviewIntervalOf(tc.expectedNextLevel)

			// todo: using plain data
			expectedNextReviewTime := data.ReviewRecords[len(data.ReviewRecords)-1].Add(
				calcIntervalByForgettingSpeed(expectedReviewInterval, calculator.userFactors.ForgettingSpeed),
			)

			// 使用 testify 的 assert 函数检查下次复习时间
			diff := nextReviewTime.Sub(expectedNextReviewTime)
			if diff < -time.Second || diff > time.Second {
				t.Errorf("Expected next review time to be within 1s of %v, but got %v (difference: %v)", expectedNextReviewTime, nextReviewTime, diff)
			}
		})
	}
}

// 基础等级调整测试
func TestCalculateReviewLevelBasic(t *testing.T) {
	// 1. 基本功能测试
	// 定义测试用例结构
	testCases := []struct {
		name            string
		inputConfidence float64
		initialLevel    int
		expectedLevel   int
	}{
		// 可以测试不同初始复习级别的边际效应
		{"None Confidence Start Level 0", ConfidenceNone, 0, 0},
		{"Low Confidence Start Level 0", ConfidenceLow, 0, 1},
		{"Medium Confidence Start Level 0", ConfidenceMedium, 0, 1},
		{"High Confidence Start Level 0", ConfidenceHigh, 0, 1},
		{"Certain Confidence Start Level 0", ConfidenceCertain, 0, 2},
		{"None Confidence Start Level 3", ConfidenceNone, 3, 0},
		{"Low Confidence Start Level 3", ConfidenceLow, 3, 1},
		{"Medium Confidence Start Level 3", ConfidenceMedium, 3, 4},
		{"High Confidence Start Level 3", ConfidenceHigh, 3, 4},
		{"Certain Confidence Start Level 3", ConfidenceCertain, 3, 6},
		{"None Confidence Start Level Max-1", ConfidenceNone, len(DefaultIntervals) - 1, 0},
		{"Low Confidence Start Level Max-1", ConfidenceLow, len(DefaultIntervals) - 1, (len(DefaultIntervals) - 1) / 2},
		{"Medium Confidence Start Level Max-1", ConfidenceMedium, len(DefaultIntervals) - 1, len(DefaultIntervals) - 3},
		{"High Confidence Start Level Max-1", ConfidenceHigh, len(DefaultIntervals) - 1, len(DefaultIntervals) - 1},
		{"Certain Confidence Start Level Max-1", ConfidenceCertain, len(DefaultIntervals) - 1, len(DefaultIntervals) - 1},
	}

	rc := &ReviewCalculator{
		reviewIntervals: DefaultIntervals,
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rd := &ReviewData{
				ReviewLevel: tc.initialLevel, // 初始复习级别由测试用例提供
			}
			newLevel := rc.CalculateReviewLevel(rd, tc.inputConfidence)
			assert.Equal(t, tc.expectedLevel, newLevel, "Expected Review Level is not as expected")
		})
	}
}

func TestAdjustForgettingSpeedWithConfidence(t *testing.T) {
	testCases := []struct {
		name            string
		forgettingSpeed float64
		setInterval     time.Duration
		realInterval    time.Duration
		confidence      float64
		expectedSpeed   float64
	}{
		{
			name:            "High confidence, longer real interval -> decrease forgetting speed",
			forgettingSpeed: 1,
			setInterval:     1 * time.Hour,
			realInterval:    2 * time.Hour,
			confidence:      ConfidenceHigh,
			expectedSpeed:   0.5,
		},
		{
			name:            "Low confidence, shorter real interval -> increase forgetting speed",
			forgettingSpeed: 1,
			setInterval:     1 * time.Hour,
			realInterval:    30 * time.Minute,
			confidence:      ConfidenceLow,
			expectedSpeed:   2,
		},
		{
			name:            "Low confidence, zero real interval -> keep forgetting speed",
			forgettingSpeed: 1,
			setInterval:     1 * time.Hour,
			realInterval:    0,
			confidence:      ConfidenceLow,
			expectedSpeed:   1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			calculator := &ReviewCalculator{
				userFactors: UserFactors{
					ForgettingSpeed: tc.forgettingSpeed,
				},
			}
			adjustedSpeed := adjustoFrgettingSpeedWithConfidence(calculator.userFactors.ForgettingSpeed, tc.setInterval, tc.realInterval, tc.confidence)
			assert.InDelta(t, tc.expectedSpeed, adjustedSpeed, 0.1, "Adjusted forgetting speed not within expected range")
		})
	}
}

func TestCalculateNextReviewWithSingleLowConfidenceReviewUsesBaseInterval(t *testing.T) {
	calculator := NewReviewCalculator(nil, UserFactors{ForgettingSpeed: 1})
	lastReview := time.Now().Add(-time.Hour)
	data := ReviewData{
		ReviewRecords: []time.Time{lastReview},
		ReviewLevel:   3,
	}

	nextReviewTime, nextReviewLevel := calculator.CalculateNextReview(&data, ConfidenceLow)

	expectedReviewInterval := calculator.ReviewIntervalOf(nextReviewLevel)
	expectedNextReviewTime := lastReview.Add(expectedReviewInterval)
	diff := nextReviewTime.Sub(expectedNextReviewTime)
	if diff < -time.Second || diff > time.Second {
		t.Errorf("CalculateNextReview(%+v, %v) time = %v, want within 1s of %v", data, ConfidenceLow, nextReviewTime, expectedNextReviewTime)
	}
	assert.Equal(t, 1.0, calculator.userFactors.ForgettingSpeed, "CalculateNextReview(%+v, %v) forgetting speed", data, ConfidenceLow)
}

func TestCalculateNextReviewWithNegativeReviewLevelAdvancesFromLevelZero(t *testing.T) {
	calculator := NewReviewCalculator(nil, UserFactors{ForgettingSpeed: 1})
	lastReview := time.Now()
	data := ReviewData{
		ReviewRecords: []time.Time{lastReview},
		ReviewLevel:   -2,
	}

	nextReviewTime, nextReviewLevel := calculator.CalculateNextReview(&data, ConfidenceHigh)

	expectedNextReviewTime := lastReview.Add(DefaultIntervals[1])
	if !nextReviewTime.Equal(expectedNextReviewTime) {
		t.Errorf("CalculateNextReview(%+v, %v) time = %v, want %v", data, ConfidenceHigh, nextReviewTime, expectedNextReviewTime)
	}
	assert.Equal(t, 1, nextReviewLevel, "CalculateNextReview(%+v, %v) review level", data, ConfidenceHigh)
}

func TestNewReviewCalculatorDefaultsEmptyIntervals(t *testing.T) {
	calculator := NewReviewCalculator([]time.Duration{}, UserFactors{ForgettingSpeed: 1})

	if got, want := calculator.MaxReviewLevel(), len(DefaultIntervals)-1; got != want {
		t.Errorf("NewReviewCalculator(empty).MaxReviewLevel() = %d, want %d", got, want)
	}
	if got, want := calculator.ReviewIntervalOf(0), DefaultIntervals[0]; got != want {
		t.Errorf("NewReviewCalculator(empty).ReviewIntervalOf(0) = %v, want %v", got, want)
	}
}

func TestCalcIntervalByForgettingSpeed(t *testing.T) {
	testCases := []struct {
		name            string
		duration        time.Duration
		speed           float64
		expectedOutcome time.Duration
	}{
		{
			name:            "Normal forgetting speed",
			duration:        1 * time.Hour,
			speed:           1,
			expectedOutcome: 1 * time.Hour,
		},
		{
			name:            "Slower forgetting speed",
			duration:        1 * time.Hour,
			speed:           0.5,
			expectedOutcome: 2 * time.Hour,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			adjustedInterval := calcIntervalByForgettingSpeed(tc.duration, tc.speed)
			assert.Equal(t, tc.expectedOutcome, adjustedInterval, "Adjusted interval does not match expected duration")
		})
	}
}
