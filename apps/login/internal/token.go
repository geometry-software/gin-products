package login

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	stdhttp "net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/geometry-software/gin-products/apps/providers/shared/http"
	"github.com/golang-jwt/jwt/v5"
)

const firebaseCertURL = "https://www.googleapis.com/robot/v1/metadata/x509/securetoken@system.gserviceaccount.com"

// TokenVerifier validates frontend-supplied Firebase ID tokens using Google's public keys.
type TokenVerifier struct {
	projectID string
	client    *stdhttp.Client
	mu        sync.Mutex
	keys      map[string]*rsa.PublicKey
	expires   time.Time
}

// NewTokenVerifier configures signature and claim validation for one Firebase project.
func NewTokenVerifier(projectID string) *TokenVerifier {
	return &TokenVerifier{projectID: projectID, client: &stdhttp.Client{Timeout: 10 * time.Second}}
}

// Verify returns trusted identity claims from a signed Firebase ID token.
func (v *TokenVerifier) Verify(ctx context.Context, raw string) (Identity, error) {
	if v.projectID == "" || raw == "" {
		return Identity{}, http.Fail(401, "Invalid Firebase ID token")
	}
	claims := jwt.MapClaims{}
	token, err := jwt.NewParser(
		jwt.WithValidMethods([]string{"RS256"}),
		jwt.WithAudience(v.projectID),
		jwt.WithIssuer("https://securetoken.google.com/"+v.projectID),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
	).ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		kid, ok := token.Header["kid"].(string)
		if !ok || kid == "" {
			return nil, http.Fail(401, "Invalid Firebase ID token")
		}
		return v.keyFor(ctx, kid)
	})
	if err != nil || !token.Valid {
		var publicError *http.Error
		if errors.As(err, &publicError) && publicError.Status == 503 {
			return Identity{}, publicError
		}
		return Identity{}, http.Fail(401, "Invalid Firebase ID token")
	}
	uid, err := claims.GetSubject()
	if err != nil || uid == "" || len(uid) > 128 {
		return Identity{}, http.Fail(401, "Invalid Firebase ID token")
	}
	iat, iatErr := claims.GetIssuedAt()
	authTime, authOK := claims["auth_time"].(float64)
	if iatErr != nil || iat == nil || !authOK || authTime < 0 || authTime > float64(time.Now().Unix()) {
		return Identity{}, http.Fail(401, "Invalid Firebase ID token")
	}
	email, _ := claims["email"].(string)
	name, _ := claims["name"].(string)
	emailVerified, _ := claims["email_verified"].(bool)
	return Identity{UID: uid, Email: email, Name: name, EmailVerified: emailVerified}, nil
}

func (v *TokenVerifier) keyFor(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if time.Now().After(v.expires) {
		if err := v.refresh(ctx); err != nil {
			return nil, err
		}
	}
	key := v.keys[kid]
	if key == nil {
		return nil, http.Fail(401, "Invalid Firebase ID token")
	}
	return key, nil
}

func (v *TokenVerifier) refresh(ctx context.Context) error {
	req, err := stdhttp.NewRequestWithContext(ctx, stdhttp.MethodGet, firebaseCertURL, nil)
	if err != nil {
		return http.Fail(503, "Firebase public keys unavailable")
	}
	resp, err := v.client.Do(req)
	if err != nil {
		return http.Fail(503, "Firebase public keys unavailable")
	}
	defer resp.Body.Close()
	if resp.StatusCode != stdhttp.StatusOK {
		return http.Fail(503, "Firebase public keys unavailable")
	}
	var certificates map[string]string
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&certificates); err != nil || len(certificates) == 0 {
		return http.Fail(503, "Invalid Firebase public keys")
	}
	keys := make(map[string]*rsa.PublicKey, len(certificates))
	for kid, encoded := range certificates {
		block, _ := pem.Decode([]byte(encoded))
		if block == nil {
			return http.Fail(503, "Invalid Firebase public keys")
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return http.Fail(503, "Invalid Firebase public keys")
		}
		key, ok := cert.PublicKey.(*rsa.PublicKey)
		if !ok {
			return http.Fail(503, "Invalid Firebase public keys")
		}
		keys[kid] = key
	}
	v.keys = keys
	v.expires = time.Now().Add(cacheMaxAge(resp.Header.Get("Cache-Control")))
	return nil
}

func cacheMaxAge(header string) time.Duration {
	for _, directive := range strings.Split(header, ",") {
		name, value, found := strings.Cut(strings.TrimSpace(directive), "=")
		if name != "max-age" || !found {
			continue
		}
		seconds, err := strconv.ParseInt(strings.Trim(value, "\""), 10, 64)
		if err == nil && seconds > 0 {
			return time.Duration(seconds) * time.Second
		}
	}
	return 0
}

var _ IdentityPort = (*TokenVerifier)(nil)
