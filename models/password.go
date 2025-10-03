package models

import (
	"time"
)

// PasswordEntry 表示一个密码条目
type PasswordEntry struct {
	ID          int       `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	Username    string    `json:"username" db:"username"`
	Password    string    `json:"password" db:"password"`
	URL         string    `json:"url" db:"url"`
	Notes       string    `json:"notes" db:"notes"`
	Category    string    `json:"category" db:"category"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Category 表示密码分类
type Category struct {
	ID   int    `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

// MasterPassword 表示主密码配置
type MasterPassword struct {
	ID           int    `json:"id" db:"id"`
	PasswordHash string `json:"password_hash" db:"password_hash"`
	Salt         string `json:"salt" db:"salt"`
}