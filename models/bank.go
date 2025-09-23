package models

type Bank struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	Name   string `gorm:"size:100;not null" json:"name"`
	Code   string `gorm:"size:20;uniqueIndex;not null" json:"code"`
	Status string `gorm:"type:enum('Active','Inactive');default:'Active'" json:"status"`
}

func (Bank) TableName() string {
	return "banks"
}
