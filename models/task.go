package models

import "time"

type Task struct {
	ID                    uint      `gorm:"primaryKey" json:"id"`
	Name                  string    `gorm:"type:varchar(100);not null" json:"name"`
	Reward                float64   `gorm:"type:decimal(15,2);not null" json:"reward"`
	RequiredLevel         int       `gorm:"not null" json:"required_level"`
	RequiredActiveMembers int64     `gorm:"not null" json:"required_active_members"`
	Status                string    `gorm:"type:enum('Active','Inactive');default:'Active'" json:"status"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type UserTask struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	TaskID    uint      `gorm:"not null" json:"task_id"`
	ClaimedAt time.Time `json:"claimed_at"`
}
