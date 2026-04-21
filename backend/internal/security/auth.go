// Package security provides JWT authentication and role-based authorization.
package security

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/AminN77/senju/backend/internal/api/problem"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const contextRoleKey = "auth.role"

// Authorizer builds Gin middleware for role-protected routes.
type Authorizer interface {
	RequireRoles(roles ...string) gin.HandlerFunc
}

// Claims are JWT claims required by the API.
type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// JWTAuthorizer validates bearer tokens and enforces route roles.
type JWTAuthorizer struct {
	secret []byte
	issuer string
	now    func() time.Time
}

// NewJWTAuthorizer creates a JWT role authorizer.
func NewJWTAuthorizer(secret, issuer string) (*JWTAuthorizer, error) {
	s := strings.TrimSpace(secret)
	if s == "" {
		return nil, errors.New("jwt secret is required")
	}
	return &JWTAuthorizer{
		secret: []byte(s),
		issuer: strings.TrimSpace(issuer),
		now:    time.Now,
	}, nil
}

// RequireRoles validates JWT bearer token and enforces at least one role match.
func (a *JWTAuthorizer) RequireRoles(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		if rr := strings.TrimSpace(strings.ToLower(r)); rr != "" {
			allowed[rr] = struct{}{}
		}
	}
	return func(c *gin.Context) {
		raw := strings.TrimSpace(c.GetHeader("Authorization"))
		if !strings.HasPrefix(strings.ToLower(raw), "bearer ") {
			unauthorized(c, "missing bearer token")
			return
		}
		token := strings.TrimSpace(raw[len("Bearer "):])
		if token == "" {
			unauthorized(c, "missing bearer token")
			return
		}
		claims := &Claims{}
		parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
			if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, errors.New("unsupported signing method")
			}
			return a.secret, nil
		})
		if err != nil || parsed == nil || !parsed.Valid {
			unauthorized(c, "invalid bearer token")
			return
		}
		if a.issuer != "" && claims.Issuer != a.issuer {
			unauthorized(c, "invalid issuer")
			return
		}
		if claims.ExpiresAt != nil && claims.ExpiresAt.Before(a.now().UTC()) {
			unauthorized(c, "token expired")
			return
		}
		role := strings.TrimSpace(strings.ToLower(claims.Role))
		if role == "" {
			forbidden(c, "missing role claim")
			return
		}
		if _, ok := allowed[role]; !ok {
			forbidden(c, "insufficient role")
			return
		}
		c.Set(contextRoleKey, role)
		c.Next()
	}
}

func unauthorized(c *gin.Context, detail string) {
	problem.JSON(c, http.StatusUnauthorized, problem.Problem{
		Type:   problem.TypeValidationError,
		Title:  "Unauthorized",
		Status: http.StatusUnauthorized,
		Detail: detail,
	})
	c.Abort()
}

func forbidden(c *gin.Context, detail string) {
	problem.JSON(c, http.StatusForbidden, problem.Problem{
		Type:   problem.TypeValidationError,
		Title:  "Forbidden",
		Status: http.StatusForbidden,
		Detail: detail,
	})
	c.Abort()
}
