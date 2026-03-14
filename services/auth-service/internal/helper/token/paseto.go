package token

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

type Claims struct {
	UserID    int64     `json:"user_id"`
	Email     string    `json:"email"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
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
		return nil, err
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

	token := paseto.NewToken()
	token.SetJti(uuid.New().String())
	token.SetIssuedAt(claims.IssuedAt)
	token.SetExpiration(claims.ExpiredAt)
	token.SetString("user_id", fmt.Sprintf("%d", userID))
	token.SetString("email", email)

	encrypted := token.V4Encrypt(m.symmetricKey, nil)
	return encrypted, claims, nil
}

func (m *Maker) VerifyToken(tokenStr string) (*Claims, error) {
	parser := paseto.NewParser()
	parser.AddRule(paseto.NotExpired())

	token, err := parser.ParseV4Local(m.symmetricKey, tokenStr, nil)
	if err != nil {
		if errors.Is(err, paseto.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	var userIDStr string
	if err := token.Get("user_id", &userIDStr); err != nil {
		return nil, ErrInvalidToken
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return nil, ErrInvalidToken
	}

	var email string
	if err := token.Get("email", &email); err != nil {
		return nil, ErrInvalidToken
	}

	return &Claims{
		UserID:    userID,
		Email:     email,
		IssuedAt:  token.GetIssuedAt(),
		ExpiredAt: token.GetExpiration(),
	}, nil
}
