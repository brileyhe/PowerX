package customerdomain

import (
	"PowerX/internal/model/powermodel"
	"time"
)

type RegisterCode struct {
	powermodel.PowerModel

	Code               string    `gorm:"comment:邀请码;unique;index" json:"code"`
	RegisterCustomerID int64     `gorm:"comment:注册客户ID" json:"registerCustomerID"`
	ExpiredAt          time.Time `gorm:"comment:到期时间" json:"expiredAt"`
}

const RegisterCodeUniqueId = "code"
