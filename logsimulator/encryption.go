package logsimulator

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	crypto_rand "crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"math/rand"

	"golang.org/x/crypto/chacha20poly1305"
)

// EncryptionType represents the available encryption algorithms
type EncryptionType string

const (
	EncryptionTypeNone     EncryptionType = "None"
	EncryptionTypeAES      EncryptionType = "AES"
	EncryptionTypeChaCha20 EncryptionType = "ChaCha20"
)

// EncryptionConfig defines the configuration for encryption simulation
type EncryptionConfig struct {
	Type       EncryptionType
	Percentage int
	AESMode    string // New field for AES mode (CBC, CTR, GCM)
	KeySize    int    // Key size in bytes (16, 24, 32 for AES)
}

// Encryptor defines the interface for encryption implementations
type Encryptor interface {
	Encrypt(plaintext string) (string, error)
	Type() EncryptionType
}

// GetEncryptor returns the appropriate encryptor based on the type and configuration
func GetEncryptor(config EncryptionConfig) (Encryptor, error) {
	switch config.Type {
	case EncryptionTypeNone:
		return &NoneEncryptor{}, nil
	case EncryptionTypeAES:
		// Default AES key size if not specified
		if config.KeySize == 0 {
			config.KeySize = 32 // Default to AES-256
		}

		// Validate key size (must be 16, 24, or 32)
		if config.KeySize != 16 && config.KeySize != 24 && config.KeySize != 32 {
			return nil, fmt.Errorf("invalid AES key size: must be 16, 24, or 32 bytes")
		}

		// Create the appropriate AES encryptor based on mode
		switch config.AESMode {
		case "CBC", "": // Default to CBC if not specified
			return NewAESCBCEncryptor(config.KeySize)
		case "CTR":
			return NewAESCTREncryptor(config.KeySize)
		case "GCM":
			return NewAESGCMEncryptor(config.KeySize)
		default:
			return nil, fmt.Errorf("unsupported AES mode: %s", config.AESMode)
		}
	case EncryptionTypeChaCha20:
		return NewChaCha20Encryptor()
	default:
		return nil, fmt.Errorf("unsupported encryption type: %s", config.Type)
	}
}

// NoneEncryptor is a pass-through implementation that does no encryption
type NoneEncryptor struct{}

func (e *NoneEncryptor) Encrypt(plaintext string) (string, error) {
	return plaintext, nil
}

func (e *NoneEncryptor) Type() EncryptionType {
	return EncryptionTypeNone
}

//-------------------- AES CBC Implementation --------------------

// AESCBCEncryptor implements AES-CBC encryption with PKCS#7 padding
type AESCBCEncryptor struct {
	key []byte
}

// NewAESCBCEncryptor creates a new AES-CBC encryptor with a random key of specified size
func NewAESCBCEncryptor(keySize int) (*AESCBCEncryptor, error) {
	key := make([]byte, keySize)
	if _, err := io.ReadFull(crypto_rand.Reader, key); err != nil {
		return nil, err
	}

	return &AESCBCEncryptor{key: key}, nil
}

func (e *AESCBCEncryptor) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	// IV needs to be unique, but not secure
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(crypto_rand.Reader, iv); err != nil {
		return "", err
	}

	// Pad plaintext to be a multiple of block size
	paddedPlaintext := padPKCS7([]byte(plaintext), aes.BlockSize)

	// Encrypt the data
	ciphertext := make([]byte, len(paddedPlaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, paddedPlaintext)

	// Prepend IV to ciphertext for decryption later
	result := append(iv, ciphertext...)

	// Base64 encode for storage
	encoded := base64.StdEncoding.EncodeToString(result)

	// Include key size in the prefix
	keyBits := len(e.key) * 8
	return fmt.Sprintf("AES-%d-CBC:%s", keyBits, encoded), nil
}

func (e *AESCBCEncryptor) Type() EncryptionType {
	return EncryptionTypeAES
}

//-------------------- AES CTR Implementation --------------------

// AESCTREncryptor implements AES-CTR (Counter Mode) encryption
type AESCTREncryptor struct {
	key []byte
}

// NewAESCTREncryptor creates a new AES-CTR encryptor with a random key of specified size
func NewAESCTREncryptor(keySize int) (*AESCTREncryptor, error) {
	key := make([]byte, keySize)
	if _, err := io.ReadFull(crypto_rand.Reader, key); err != nil {
		return nil, err
	}

	return &AESCTREncryptor{key: key}, nil
}

func (e *AESCTREncryptor) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(crypto_rand.Reader, iv); err != nil {
		return "", err
	}

	// CTR mode doesn't require padding, but does need a counter, which is
	// stored in the IV.
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(plaintext))

	// Base64 encode for storage
	encoded := base64.StdEncoding.EncodeToString(ciphertext)

	// Include key size in the prefix
	keyBits := len(e.key) * 8
	return fmt.Sprintf("AES-%d-CTR:%s", keyBits, encoded), nil
}

func (e *AESCTREncryptor) Type() EncryptionType {
	return EncryptionTypeAES
}

//-------------------- AES GCM Implementation --------------------

// AESGCMEncryptor implements AES-GCM (Galois/Counter Mode) authenticated encryption
type AESGCMEncryptor struct {
	key []byte
}

// NewAESGCMEncryptor creates a new AES-GCM encryptor with a random key of specified size
func NewAESGCMEncryptor(keySize int) (*AESGCMEncryptor, error) {
	key := make([]byte, keySize)
	if _, err := io.ReadFull(crypto_rand.Reader, key); err != nil {
		return nil, err
	}

	return &AESGCMEncryptor{key: key}, nil
}

func (e *AESGCMEncryptor) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	// Create a new GCM cipher
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Create a nonce
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(crypto_rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt and seal the data
	// GCM produces ciphertext with authentication tag appended
	ciphertext := aead.Seal(nil, nonce, []byte(plaintext), nil)

	// Prepend nonce to ciphertext
	result := append(nonce, ciphertext...)

	// Base64 encode for storage
	encoded := base64.StdEncoding.EncodeToString(result)

	// Include key size in the prefix
	keyBits := len(e.key) * 8
	return fmt.Sprintf("AES-%d-GCM:%s", keyBits, encoded), nil
}

func (e *AESGCMEncryptor) Type() EncryptionType {
	return EncryptionTypeAES
}

//-------------------- ChaCha20 Implementation --------------------

// ChaCha20Encryptor implements ChaCha20-Poly1305 encryption
type ChaCha20Encryptor struct {
	key []byte
}

// NewChaCha20Encryptor creates a new ChaCha20 encryptor with a random key
func NewChaCha20Encryptor() (*ChaCha20Encryptor, error) {
	// Generate a random 32-byte key for ChaCha20-Poly1305
	key := make([]byte, chacha20poly1305.KeySize)
	if _, err := io.ReadFull(crypto_rand.Reader, key); err != nil {
		return nil, err
	}

	return &ChaCha20Encryptor{key: key}, nil
}

func (e *ChaCha20Encryptor) Encrypt(plaintext string) (string, error) {
	aead, err := chacha20poly1305.New(e.key)
	if err != nil {
		return "", err
	}

	// Generate a random nonce
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(crypto_rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt the plaintext
	ciphertext := aead.Seal(nil, nonce, []byte(plaintext), nil)

	// Prepend nonce to ciphertext for decryption later
	result := append(nonce, ciphertext...)

	// Base64 encode for storage
	encoded := base64.StdEncoding.EncodeToString(result)
	return fmt.Sprintf("ChaCha20:%s", encoded), nil
}

func (e *ChaCha20Encryptor) Type() EncryptionType {
	return EncryptionTypeChaCha20
}

//-------------------- Helper Functions --------------------

// padPKCS7 pads data to a multiple of blockSize according to PKCS#7
func padPKCS7(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

// MaybeEncrypt encrypts a value based on encryption configuration and random chance
func MaybeEncrypt(value string, config EncryptionConfig) (string, error) {
	// If encryption is disabled or percentage is 0, return the original value
	if config.Type == EncryptionTypeNone || config.Percentage <= 0 {
		return value, nil
	}

	// Check if we should encrypt this value based on the percentage
	if config.Percentage < 100 && (rand.Intn(100) >= config.Percentage) {
		return value, nil
	}

	// Get the appropriate encryptor
	enc, err := GetEncryptor(config)
	if err != nil {
		return value, err
	}

	// Encrypt the value
	return enc.Encrypt(value)
}
