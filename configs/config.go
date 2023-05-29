package configs

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"
)

type Server struct {
	Name     string `json:"name"`
	Port     int    `json:"port"`
	KeysPath string `json:"keys_path"`
}

type Database struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type JWT struct {
	KeyID            string `json:"key_id"`
	SigningAlgorithm string `json:"signing_algorithm"`
	Issuer           string `json:"issuer"`
	ExpiresIn        string `json:"expires_in"`
	RefreshThreshold string `json:"refresh_threshold"`
	Source           string `json:"source"`
}

type WhatsApp struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

func (j *JWT) IsSourceCookies() bool {
	if j.Source == "" {
		return true
	}
	return j.Source == "cookies"
}

func (j *JWT) GetExpiresIn() time.Duration {
	duration, err := time.ParseDuration(j.ExpiresIn)
	if err != nil {
		return time.Minute * 30
	}
	return duration
}

func (j *JWT) GetRefreshThreshold() time.Duration {
	duration, err := time.ParseDuration(j.RefreshThreshold)
	if err != nil {
		return time.Minute * 30
	}
	return duration
}

type KeyPair struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
}

type Config struct {
	Server   Server   `json:"server"`
	Database Database `json:"db"`
	JWT      JWT      `json:"jwt"`
	WhatsApp WhatsApp `json:"wpp"`
	Keys     Keys
}

type Keys struct {
	JWT KeyPair
	TLS *tls.Certificate
}

var instance *Config

func Load(settingsPath string) error {
	if instance != nil {
		return nil
	}
	file, err := os.Open(settingsPath)
	if err != nil {
		return err
	}
	defer file.Close()
	configs := &Config{}
	err = json.NewDecoder(file).Decode(configs)
	if err != nil {
		return err
	}

	err = configs.loadKeys()
	if err != nil {
		return err
	}
	instance = configs
	return nil
}

func Get() *Config {
	return instance
}

func (c *Config) loadKeys() error {
	jwtKeys, err := c.loadKey("jwt")
	if err != nil {
		return err
	}
	c.Keys.JWT = *jwtKeys

	tlsKeys, err := c.loadTLSKey()
	if err != nil {
		return err
	}
	c.Keys.TLS = tlsKeys
	return nil
}

func (c *Config) loadTLSKey() (*tls.Certificate, error) {
	certPath := fmt.Sprintf("%s/%s", c.Server.KeysPath, "fullchain.pem")
	keyPath := fmt.Sprintf("%s/%s", c.Server.KeysPath, "privkey.pem")
	// Try to load the certificate and key from the files
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err == nil {
		log.Printf("Loaded certificate and key from files: %s, %s", certPath, keyPath)
		return &cert, nil
	}

	// Generate a new certificate and key
	log.Printf("Failed to load certificate and key from files: %s, %s. Generating new ones...", certPath, keyPath)

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %v", err)
	}

	template := x509.Certificate{
		SerialNumber:          bigInt(1),
		Subject:               pkixName(c.Server.Name),
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %v", err)
	}

	certOut, err := os.Create(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s for writing: %v", certPath, err)
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return nil, fmt.Errorf("failed to write data to %s: %v", certPath, err)
	}
	if err := certOut.Close(); err != nil {
		return nil, fmt.Errorf("error closing %s: %v", certPath, err)
	}
	log.Printf("Certificate written to file: %s", certPath)

	keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s for writing: %v", keyPath, err)
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %v", err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return nil, fmt.Errorf("failed to write data to %s: %v", keyPath, err)
	}
	if err := keyOut.Close(); err != nil {
		return nil, fmt.Errorf("error closing %s: %v", keyPath, err)
	}
	log.Printf("Key written to file: %s", keyPath)

	// Load the newly generated certificate and key
	return c.loadTLSKey()
}

func bigInt(i int64) *big.Int {
	return big.NewInt(i)
}

func pkixName(commonName string) pkix.Name {
	return pkix.Name{
		CommonName: commonName,
	}
}

func (c *Config) generateKeys(keyName string) (*KeyPair, error) {
	fmt.Printf("generating %s keys...\n", keyName)
	// Generate private key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	// Encode private key to PEM format
	privateKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	privateKeyPem := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	// Write private key to file
	privateKeyPath := fmt.Sprintf("%s/%s.pem", c.Server.KeysPath, keyName)
	privateKeyFile, err := os.Create(privateKeyPath)
	if err != nil {
		return nil, err
	}
	defer privateKeyFile.Close()
	if err := pem.Encode(privateKeyFile, privateKeyPem); err != nil {
		return nil, err
	}

	// Generate public key from private key
	publicKey := privateKey.PublicKey

	// Encode public key to PEM format
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&publicKey)
	if err != nil {
		return nil, err
	}
	publicKeyPem := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	// Write public key to file
	publicKeyPath := fmt.Sprintf("%s/%s.pub.pem", c.Server.KeysPath, keyName)
	publicKeyFile, err := os.Create(publicKeyPath)
	if err != nil {
		return nil, err
	}
	defer publicKeyFile.Close()
	if err := pem.Encode(publicKeyFile, publicKeyPem); err != nil {
		return nil, err
	}
	return &KeyPair{
		PrivateKey: privateKey,
		PublicKey:  &publicKey,
	}, nil
}

func (c *Config) loadKey(keyName string) (*KeyPair, error) {
	path := fmt.Sprintf("%s/%s.pem", c.Server.KeysPath, keyName)
	keyBytes, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("failed to read %s private key file: %v\n", keyName, err)
		return c.generateKeys(keyName)
	}

	// Decode the PEM-encoded private key
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		fmt.Printf("failed to decode %s private key\n", keyName)
		return c.generateKeys(keyName)
	}

	// Parse the DER-encoded RSA private key
	privKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		fmt.Printf("failed to parse %s private key: %v\n", keyName, err)
		return c.generateKeys(keyName)
	}

	return &KeyPair{
		PrivateKey: privKey,
		PublicKey:  &privKey.PublicKey,
	}, nil
}
