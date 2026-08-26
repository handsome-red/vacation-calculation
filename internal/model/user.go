// internal/model/vacation.go
package model

import "time"

type User struct {
	ID         uint      `json:"id" gorm:"primary_key"`
	LastName   string    `json:"last_name" gorm:"not null;size:100"`
	FirstName  string    `json:"first_name" gorm:"not null;size:100"`
	MiddleName string    `json:"middle_name,omitempty" gorm:"size:100"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (u *User) FullName() string {
	return u.LastName + u.FirstName + u.MiddleName
}
