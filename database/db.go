package database

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"hank.com/password_tool/crypto"
	"hank.com/password_tool/models"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
	key  []byte
}

// NewDB 创建新的数据库连接
func NewDB() (*DB, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dbDir := filepath.Join(homeDir, ".password_tool")
	if err := os.MkdirAll(dbDir, 0700); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(dbDir, "passwords.db")
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	db := &DB{conn: conn}
	if err := db.createTables(); err != nil {
		return nil, err
	}

	return db, nil
}

// SetMasterKey 设置主密钥
func (db *DB) SetMasterKey(key []byte) {
	db.key = key
}

// createTables 创建数据库表
func (db *DB) createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS master_password (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			password_hash TEXT NOT NULL,
			salt TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS password_entries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			username TEXT,
			password TEXT NOT NULL,
			url TEXT,
			notes TEXT,
			category TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, query := range queries {
		if _, err := db.conn.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

// SetMasterPassword 设置主密码
func (db *DB) SetMasterPassword(password string) error {
	salt, err := crypto.GenerateSalt()
	if err != nil {
		return err
	}

	hash := crypto.HashMasterPassword(password, salt)
	saltStr := base64.StdEncoding.EncodeToString(salt)

	_, err = db.conn.Exec("INSERT OR REPLACE INTO master_password (id, password_hash, salt) VALUES (1, ?, ?)", hash, saltStr)
	return err
}

// VerifyMasterPassword 验证主密码
func (db *DB) VerifyMasterPassword(password string) (bool, error) {
	var hash, saltStr string
	err := db.conn.QueryRow("SELECT password_hash, salt FROM master_password WHERE id = 1").Scan(&hash, &saltStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	salt, err := base64.StdEncoding.DecodeString(saltStr)
	if err != nil {
		return false, err
	}

	return crypto.VerifyMasterPassword(password, hash, salt), nil
}

// GetMasterPasswordSalt 获取主密码盐值
func (db *DB) GetMasterPasswordSalt() ([]byte, error) {
	var saltStr string
	err := db.conn.QueryRow("SELECT salt FROM master_password WHERE id = 1").Scan(&saltStr)
	if err != nil {
		return nil, err
	}

	return base64.StdEncoding.DecodeString(saltStr)
}

// HasMasterPassword 检查是否已设置主密码
func (db *DB) HasMasterPassword() (bool, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM master_password").Scan(&count)
	return count > 0, err
}

// AddPasswordEntry 添加密码条目
func (db *DB) AddPasswordEntry(entry *models.PasswordEntry) error {
	if db.key == nil {
		return fmt.Errorf("master key not set")
	}

	encryptedPassword, err := crypto.Encrypt([]byte(entry.Password), db.key)
	if err != nil {
		return err
	}

	now := time.Now()
	_, err = db.conn.Exec(`
		INSERT INTO password_entries (title, username, password, url, notes, category, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		entry.Title, entry.Username, encryptedPassword, entry.URL, entry.Notes, entry.Category, now, now)

	return err
}

// GetPasswordEntries 获取所有密码条目
func (db *DB) GetPasswordEntries() ([]*models.PasswordEntry, error) {
	if db.key == nil {
		return nil, fmt.Errorf("master key not set")
	}

	rows, err := db.conn.Query(`
		SELECT id, title, username, password, url, notes, category, created_at, updated_at
		FROM password_entries ORDER BY title`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*models.PasswordEntry
	for rows.Next() {
		entry := &models.PasswordEntry{}
		var encryptedPassword string
		err := rows.Scan(&entry.ID, &entry.Title, &entry.Username, &encryptedPassword,
			&entry.URL, &entry.Notes, &entry.Category, &entry.CreatedAt, &entry.UpdatedAt)
		if err != nil {
			return nil, err
		}

		decryptedPassword, err := crypto.Decrypt(encryptedPassword, db.key)
		if err != nil {
			return nil, err
		}
		entry.Password = string(decryptedPassword)

		entries = append(entries, entry)
	}

	return entries, nil
}

// UpdatePasswordEntry 更新密码条目
func (db *DB) UpdatePasswordEntry(entry *models.PasswordEntry) error {
	if db.key == nil {
		return fmt.Errorf("master key not set")
	}

	encryptedPassword, err := crypto.Encrypt([]byte(entry.Password), db.key)
	if err != nil {
		return err
	}

	_, err = db.conn.Exec(`
		UPDATE password_entries 
		SET title=?, username=?, password=?, url=?, notes=?, category=?, updated_at=?
		WHERE id=?`,
		entry.Title, entry.Username, encryptedPassword, entry.URL, entry.Notes, entry.Category, time.Now(), entry.ID)

	return err
}

// DeletePasswordEntry 删除密码条目
func (db *DB) DeletePasswordEntry(id int) error {
	_, err := db.conn.Exec("DELETE FROM password_entries WHERE id=?", id)
	return err
}

// GetCategories 获取所有分类
func (db *DB) GetCategories() ([]*models.Category, error) {
	rows, err := db.conn.Query("SELECT id, name FROM categories ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*models.Category
	for rows.Next() {
		category := &models.Category{}
		err := rows.Scan(&category.ID, &category.Name)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

// AddCategory 添加分类
func (db *DB) AddCategory(name string) error {
	_, err := db.conn.Exec("INSERT INTO categories (name) VALUES (?)", name)
	return err
}

// Close 关闭数据库连接
func (db *DB) Close() error {
	return db.conn.Close()
}