package model

import (
	"time"

	"github.com/bagaking/memorianexus/internal/utils"

	"gorm.io/gorm"
)

// ProfilePoints 定义了用户积分信息的模型
type ProfilePoints struct {
	ID        utils.UInt64   `gorm:"primaryKey;autoIncrement:false"` // 与用户ID一致
	Cash      utils.UInt64   `gorm:"default:0"`                      // 现金
	Gem       utils.UInt64   `gorm:"default:0"`                      // 宝石
	VipScore  utils.UInt64   `gorm:"default:0"`                      // VIP 积分
	CreatedAt time.Time      // 记录的创建时间
	UpdatedAt time.Time      // 记录的更新时间
	DeletedAt gorm.DeletedAt `gorm:"index"` // 记录的删除时间
}

// TableName 自定义表名
func (ProfilePoints) TableName() string {
	return "profile_points"
}

// EnsureLoadProfilePoints 从数据库中"懒加载"用户高级设置
func (p *Profile) EnsureLoadProfilePoints(db *gorm.DB) (*ProfilePoints, error) {
	if p.points != nil {
		return p.points, nil
	}
	return EnsureLoadProfilePoints(db, p.ID)
}

// EnsureLoadProfilePoints 从数据库中加载用户积分信息
func EnsureLoadProfilePoints(db *gorm.DB, userID utils.UInt64) (*ProfilePoints, error) {
	points := &ProfilePoints{ID: userID}
	result := db.FirstOrCreate(points, points)
	if result.Error != nil {
		return nil, result.Error
	}

	return points, nil
}

// AddUserCash 更新用户积分
func AddUserCash(db *gorm.DB, userID utils.UInt64, points int) error {
	return db.Model(&ProfilePoints{}).Where("id = ?", userID).
		Update("cash", gorm.Expr("cash + ?", points)).
		Error
}

// todo: update
