package security

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func TestRequireRoles(t *testing.T) {
	t.Parallel()
	auth, err := NewJWTAuthorizer("secret", "senju")
	if err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	r.GET("/protected", auth.RequireRoles("runner", "admin"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	t.Run("missing token", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("status=%d", w.Code)
		}
	})

	t.Run("wrong role", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+signedToken(t, "secret", "senju", "analyst"))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("status=%d", w.Code)
		}
	})

	t.Run("allowed role", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+signedToken(t, "secret", "senju", "runner"))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("status=%d", w.Code)
		}
	})
}

func signedToken(t *testing.T, secret, issuer, role string) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   "test-user",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	})
	out, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	return out
}
