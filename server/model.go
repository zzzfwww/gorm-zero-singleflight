package server

import (
	"time"
)

type ThirdPayment struct {
	Id             int64     `json:"id" gorm:"primaryKey"`
	Sn             string    `json:"sn"` // 流水单号
	CreateTime     time.Time `json:"create_time"`
	UpdateTime     time.Time `json:"update_time"`
	DeleteTime     time.Time `json:"delete_time"`
	DelState       int64     `json:"del_state"`
	Version        int64     `json:"version"`          // 乐观锁版本号
	UserId         int64     `json:"user_id"`          // 用户id
	PayMode        string    `json:"pay_mode"`         // 支付方式 1:微信支付
	TradeType      string    `json:"trade_type"`       // 第三方支付类型
	TradeState     string    `json:"trade_state"`      // 第三方交易状态
	PayTotal       int64     `json:"pay_total"`        // 支付总金额(分)
	TransactionId  string    `json:"transaction_id"`   // 第三方支付单号
	TradeStateDesc string    `json:"trade_state_desc"` // 支付状态描述
	OrderSn        string    `json:"order_sn"`         // 业务单号
	ServiceType    string    `json:"service_type"`     // 业务类型
	PayStatus      int64     `json:"pay_status"`       // 平台内交易状态   -1:支付失败 0:未支付 1:支付成功 2:已退款
	PayTime        time.Time `json:"pay_time"`         // 支付成功时间
}

func (ThirdPayment) TableName() string {
	return "third_payment"
}
