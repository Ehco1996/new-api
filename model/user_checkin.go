package model

import (
	"errors"
	"fmt"
	"math/rand"
	"one-api/common"
	"time"

	"gorm.io/gorm"
)

const (
	TokenToDollar = 500000
)

type UserCheckInLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    int       `gorm:"index;not null;column:user_id" json:"user_id"`       // 关联到User模型，明确指定列名为user_id
	Date      time.Time `gorm:"uniqueIndex:user_date_idx;column:date" json:"date"`  // 确保同一个用户每天只能有一条记录，明确指定列名为date
	GiftQuota int       `gorm:"not null;column:gift_quota" json:"gift_quota"`       // 签到赠送的随机quota值，明确指定列名为gift_quota
	CreatedAt time.Time `gorm:"autoCreateTime;column:created_at" json:"created_at"` // 记录创建时间，明确指定列名为created_at
}

func (UserCheckInLog) TableName() string {
	return "user_check_in_logs"
}

func getRandomQuota() (int, error) {
	minDollar, err := common.GetIntEnv("MIN_CHECKIN_DOLLAR")
	if err != nil {
		return 0, err

	}
	maxDollar, err := common.GetIntEnv("MAX_CHECKIN_DOLLAR")
	if err != nil {
		return 0, err
	}
	if minDollar >= maxDollar {
		return 0, fmt.Errorf("MIN_CHECKIN_DOLLAR 必须小于 MAX_CHECKIN_DOLLAR")
	}
	randomQuota := rand.Intn(maxDollar-minDollar) + minDollar
	randomQuota *= TokenToDollar
	return randomQuota, nil
}

// CheckIn 用户签到功能
func CheckIn(userID int) (*UserCheckInLog, error) {
	var checkInLog *UserCheckInLog
	// 确保在一个事务中执行
	err := DB.Transaction(func(tx *gorm.DB) error {
		today := time.Now().Format("2006-01-02") // 获取今天的日期字符串
		var count int64
		// 检查用户今天是否已经签到
		if err := tx.Model(&UserCheckInLog{}).Where("user_id = ? AND DATE(date) = ?", userID, today).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			// 如果已经签到过，返回错误
			return errors.New("今日已签到，请明天再来")
		}

		// 生成随机GiftQuota值，比如1到10之间
		giftQuota, err := getRandomQuota()
		if err != nil {
			return err
		}
		// 增加用户Quota
		if err := IncreaseUserQuotaWithTX(tx, userID, giftQuota); err != nil {
			return err
		}

		// 创建签到记录
		checkInLog = &UserCheckInLog{
			UserID:    userID,
			Date:      time.Now(),
			GiftQuota: giftQuota,
		}
		if err := tx.Create(checkInLog).Error; err != nil {
			return err
		}

		RecordLog(checkInLog.UserID, LogTypeTopup,
			fmt.Sprintf("签到赠送金额: %v", common.LogQuota(checkInLog.GiftQuota)))
		return nil
	})

	// 如果事务执行成功，返回签到记录和nil错误；否则返回nil和相应的错误信息
	if err != nil {
		return nil, err
	}
	return checkInLog, nil
}
