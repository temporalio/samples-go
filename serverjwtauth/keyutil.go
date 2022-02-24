package serverjwtauth

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

var keyFile string

func init() {
	_, currFile, _, _ := runtime.Caller(0)
	keyFile = filepath.Join(currFile, "../key.priv.pem")
}

func GenAndWriteKey() error {
	// We'll just generate an ECDSA P-256 priv key and write to disk. Since this
	// is a sample we will not encrypt the private key, but in real-world cases
	// you would (or get it from an external system).
	log.Print("Generating key")
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("failed generating key: %w", err)
	}

	// Write PEM private key
	b, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return fmt.Errorf("failed marshalling key: %w", err)
	}
	b = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: b})
	if err := os.WriteFile(keyFile, b, 0600); err != nil {
		return fmt.Errorf("failed writing key: %w", err)
	}

	log.Print("Key generated")
	return nil
}

func ReadKey() (*ecdsa.PrivateKey, *jose.JSONWebKey, error) {
	b, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed reading key file: %w", err)
	}
	block, _ := pem.Decode(b)
	if block == nil || block.Type != "PRIVATE KEY" {
		return nil, nil, fmt.Errorf("invalid key PEM")
	}
	keyIface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed parsing private key: %w", err)
	}
	key, _ := keyIface.(*ecdsa.PrivateKey)
	if key == nil {
		return nil, nil, fmt.Errorf("private key not ECDSA key, got type: %T", key)
	}

	jwk := &jose.JSONWebKey{Key: key.Public(), Algorithm: string(jose.ES256), Use: "sig"}
	// Make the key ID the URL-encoded thumbprint, but anything can be used
	b, err = jwk.Thumbprint(crypto.SHA256)
	if err != nil {
		return nil, nil, fmt.Errorf("failed generating thumbprint: %w", err)
	}
	jwk.KeyID = base64.URLEncoding.EncodeToString(b)

	return key, jwk, nil
}

type JWTConfig struct {
	Key         *ecdsa.PrivateKey
	KeyID       string
	Permissions []string
	// "exp" is overridden and "sub" is defaulted
	ClaimsTemplate jwt.Claims
	ExtraClaims    interface{}
	// Default is 10m
	Expiration time.Duration
}

type claims struct {
	jwt.Claims
	Permissions []string `json:"permissions,omitempty"`
}

func (j *JWTConfig) GenToken() (string, error) {
	sig, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.ES256, Key: j.Key},
		(&jose.SignerOptions{}).WithType("JWT").WithHeader("kid", j.KeyID),
	)
	if err != nil {
		return "", fmt.Errorf("failed creating signer: %w", err)
	}
	claims := &claims{Claims: j.ClaimsTemplate}
	if claims.Subject == "" {
		claims.Subject = "temporal-samples-go"
	}

	// Set expiration
	expiration := j.Expiration
	if expiration == 0 {
		expiration = 10 * time.Minute
	}
	claims.Expiry = jwt.NewNumericDate(time.Now().Add(expiration))

	// Set permissions and key ID
	claims.Permissions = j.Permissions

	// Gen
	builder := jwt.Signed(sig).Claims(claims)
	if j.ExtraClaims != nil {
		builder = builder.Claims(j.ExtraClaims)
	}
	token, err := builder.CompactSerialize()
	if err != nil {
		return "", fmt.Errorf("failed serializing token: %w", err)
	}
	return token, nil
}

type JWTHeadersProvider struct {
	Config JWTConfig

	// How long before expiration before regen
	RegenLeeway time.Duration

	reuseToken      string
	reuseRegenAfter time.Time
	reuseLock       sync.RWMutex
}

func (j *JWTHeadersProvider) GetHeaders(ctx context.Context) (map[string]string, error) {
	// See if there is a token that can be reused
	j.reuseLock.RLock()
	token := j.reuseToken
	regenAfter := j.reuseRegenAfter
	j.reuseLock.RUnlock()

	// If regen after has passed (which it will if this is first use), we want to
	// regen
	if time.Now().After(regenAfter) {
		var err error
		// We intentionally don't regen under lock
		if token, err = j.Config.GenToken(); err != nil {
			return nil, err
		}
		expiration := j.Config.Expiration
		if expiration == 0 {
			expiration = 10 * time.Minute
		}
		regenLeeway := j.RegenLeeway
		if regenLeeway == 0 {
			regenLeeway = 1 * time.Minute
		}
		regenAfter = time.Now().Add(expiration - regenLeeway)
		// Set the token during lock. In racy situations, we can set this multiple
		// times unnecessarily when concurrently checking previous regen. We accept
		// this for simplicity over a compare-and-swap loop or long lock.
		j.reuseLock.Lock()
		j.reuseToken = token
		j.reuseRegenAfter = regenAfter
		j.reuseLock.Unlock()
	}

	// Return header
	return map[string]string{"Authorization": "Bearer " + token}, nil
}
