package models

import "time"

// ExperimentTask groups reservations submitted atomically for a sequential experiment.
type ExperimentTask struct {
	ID            uint          `gorm:"primaryKey" json:"id"`
	UserID        uint          `gorm:"not null;index" json:"user_id"`
	Name          string        `gorm:"size:100;not null" json:"name"`
	Purpose       string        `gorm:"type:text" json:"purpose"`
	EarliestStart time.Time     `json:"earliest_start"`
	Deadline      time.Time     `json:"deadline"`
	CreatedAt     time.Time     `json:"created_at"`
	Reservations  []Reservation `gorm:"foreignKey:TaskID" json:"reservations"`
}
