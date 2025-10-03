package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

const (
	saltSize = 32
	keySize  = 32
	nonceSize = 12
)

// GenerateSalt 生成随机盐值
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, saltSize)
	_, err := rand.Read(salt)
	return salt, err
}

// DeriveKey 从主密码和盐值派生加密密钥
func DeriveKey(masterPassword string, salt []byte) []byte {
	return pbkdf2.Key([]byte(masterPassword), salt, 100000, keySize, sha256.New)
}

// HashMasterPassword 对主密码进行哈希处理
func HashMasterPassword(masterPassword string, salt []byte) string {
	hash := pbkdf2.Key([]byte(masterPassword), salt, 100000, 32, sha256.New)
	return base64.StdEncoding.EncodeToString(hash)
}

// Encrypt 使用AES-GCM加密数据
func Encrypt(plaintext, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 使用AES-GCM解密数据
func Decrypt(ciphertext string, key []byte) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, err
	}

	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := data[:nonceSize]
	ciphertext_bytes := data[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext_bytes, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// VerifyMasterPassword 验证主密码
func VerifyMasterPassword(inputPassword, storedHash string, salt []byte) bool {
	inputHash := HashMasterPassword(inputPassword, salt)
	return inputHash == storedHash
}