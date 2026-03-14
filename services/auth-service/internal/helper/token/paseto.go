package token

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/google/uuid"

	"github.com/bagasss3/toko/services/auth-service/internal/model"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

type Claims struct {
	UserID    uuid.UUID
	Email     string
	Role      model.Role
	IssuedAt  time.Time
	ExpiredAt time.Time
}

type Maker struct {
	symmetricKey paseto.V4SymmetricKey
}

func NewMaker(symmetricKey string) (*Maker, error) {
	if len(symmetricKey) != 32 {
		return nil, errors.New("symmetric key must be exactly 32 characters")
	}

	key, err := paseto.V4SymmetricKeyFromBytes([]byte(symmetricKey))
	if err != nil {
		return nil, fmt.Errorf("creating symmetric key: %w", err)
	}

	return &Maker{symmetricKey: key}, nil
}

func (m *Maker) CreateToken(userID uuid.UUID, email string, role model.Role, duration time.Duration) (string, *Claims, error) {
	claims := &Claims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}

	t := paseto.NewToken()
	t.SetJti(uuid.New().String())
	t.SetIssuedAt(claims.IssuedAt)
	t.SetExpiration(claims.ExpiredAt)
	t.SetString("user_id", userID.String())
	t.SetString("email", email)
	t.SetString("role", string(role))

	encrypted := t.V4Encrypt(m.symmetricKey, nil)
	return encrypted, claims, nil
}

func (m *Maker) VerifyToken(tokenStr string) (*Claims, error) {
	parser := paseto.NewParser()
	parser.AddRule(paseto.NotExpired())

	t, err := parser.ParseV4Local(m.symmetricKey, tokenStr, nil)
	if err != nil {
		if strings.Contains(err.Error(), "token has expired") {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	userIDStr, err := t.GetString("user_id")
	if err != nil {
		return nil, ErrInvalidToken
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, ErrInvalidToken
	}

	email, err := t.GetString("email")
	if err != nil {
		return nil, ErrInvalidToken
	}

	roleStr, err := t.GetString("role")
	if err != nil {
		return nil, ErrInvalidToken
	}

	issuedAt, err := t.GetIssuedAt()
	if err != nil {
		return nil, ErrInvalidToken
	}

	expiredAt, err := t.GetExpiration()
	if err != nil {
		return nil, ErrInvalidToken
	}

	return &Claims{
		UserID:    userID,
		Email:     email,
		Role:      model.Role(roleStr),
		IssuedAt:  issuedAt,
		ExpiredAt: expiredAt,
	}, nil
}
