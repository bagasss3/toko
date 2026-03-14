package token

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

type Claims struct {
	UserID    int64
	Email     string
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

func (m *Maker) CreateToken(userID int64, email string, duration time.Duration) (string, *Claims, error) {
	claims := &Claims{
		UserID:    userID,
		Email:     email,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}

	t := paseto.NewToken()
	t.SetJti(uuid.New().String())
	t.SetIssuedAt(claims.IssuedAt)
	t.SetExpiration(claims.ExpiredAt)
	t.SetString("user_id", strconv.FormatInt(userID, 10))
	t.SetString("email", email)

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

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return nil, ErrInvalidToken
	}

	email, err := t.GetString("email")
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
		IssuedAt:  issuedAt,
		ExpiredAt: expiredAt,
	}, nil
}
