package metaaccess

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
)

func DecryptionPin(Encrypted []byte, prikey string, creatorPubkey string, encryptedKey string) (result []byte, err error) {
	//get sp
	var rePrikey *ecdh.PrivateKey
	rePrikey, err = BuildPrivateKey(prikey)
	if err != nil {
		return
	}
	var rePubkey *ecdh.PublicKey
	rePubkey, err = BuildPublicKey(creatorPubkey)
	if err != nil {
		return
	}
	var sp []byte
	sp, err = PerformECDH(rePrikey, rePubkey)
	if err != nil {
		return
	}
	hash := sha256.Sum256(sp)
	spStr := hex.EncodeToString(hash[:])
	spb, _ := hex.DecodeString(spStr)
	// decrypt p1
	//var p1b []byte
	p1b, err := hex.DecodeString(encryptedKey)
	if err != nil {
		return
	}
	decryptedP1, err := DecryptPayloadAES(spb, p1b)
	if err != nil {
		return
	}
	encontentb, err := hex.DecodeString(string(Encrypted))
	if err != nil {
		return
	}
	result, err = DecryptPayloadAES(decryptedP1, encontentb)
	return
}
func GenerateAESKey() (string, error) {
	key := make([]byte, 32) // AES-256 key
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	return hexEncode(key), nil
}
func hexEncode(data []byte) (hexstring string) {
	return hex.EncodeToString(data)
}
func hexDecode(str string) (data []byte, err error) {
	return hex.DecodeString(str)
}
func GenKeyPair() (privateKey string, publicKey string, e error) {
	curve := ecdh.P256()
	// A 生成自己的私钥
	privKeyA, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		return
	}
	privateKey = hex.EncodeToString(privKeyA.Bytes())
	publicKey = hex.EncodeToString(privKeyA.PublicKey().Bytes())
	return
}
func GetAesContent(content string, p1 string) (aesContent string, err error) {
	payload := []byte(content)
	// 使用 P1 加密 Payload
	P1, err := hexDecode(p1)
	if err != nil {
		return
	}
	encryptedPayload, err := EncryptPayloadAES(P1, payload)
	if err != nil {
		return
	}
	aesContent = hexEncode(encryptedPayload)
	return
}

// EncryptPayloadAES encrypts payload using AES
func EncryptPayloadAES(key, payload []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(payload))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], payload)

	return ciphertext, nil
}
func BuildPrivateKey(privateKeyStr string) (priKey *ecdh.PrivateKey, err error) {
	decodedKeyBytes, err := hex.DecodeString(privateKeyStr)
	if err != nil {
		return
	}
	curve := ecdh.P256()
	priKey, err = curve.NewPrivateKey(decodedKeyBytes)
	return
}

func BuildPublicKey(publicKeyStr string) (pubKey *ecdh.PublicKey, err error) {
	decodedPubKeyBytes, err := hex.DecodeString(publicKeyStr)
	if err != nil {
		return
	}
	curve := ecdh.P256()
	pubKey, err = curve.NewPublicKey(decodedPubKeyBytes)
	return
}

func PerformECDH(privKeyA *ecdh.PrivateKey, pubKeyB *ecdh.PublicKey) ([]byte, error) {
	return privKeyA.ECDH(pubKeyB)
}

// DecryptPayloadAES decrypts payload using AES
func DecryptPayloadAES(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return ciphertext, nil
}
