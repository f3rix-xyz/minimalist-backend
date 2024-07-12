package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name                  string
	Phone                 string
	SubscriptionValidTill time.Time
}
