package application

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"jurnal-be/internal/domain"
	"jurnal-be/internal/infrastructure/authutil"
	"jurnal-be/internal/infrastructure/oauth"
)

var ErrInvalidCredentials = errors.New("email atau password salah")

type AuthService struct {
	users         domain.UserRepository
	refreshTokens domain.RefreshTokenRepository
	googleAud     string
	accessSecret  string
	refreshSecret string
}

func NewAuthService(users domain.UserRepository, rt domain.RefreshTokenRepository, googleAud, accessSecret, refreshSecret string) *AuthService {
	return &AuthService{
		users:         users,
		refreshTokens: rt,
		googleAud:     googleAud,
		accessSecret:  accessSecret,
		refreshSecret: refreshSecret,
	}
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// LoginWithGoogle: verifikasi ID token ke Google, cari atau bikin user baru, lalu keluarkan token pair.
func (s *AuthService) LoginWithGoogle(ctx context.Context, idToken, deviceInfo string) (*TokenPair, error) {
	profile, err := oauth.VerifyIDToken(ctx, idToken, s.googleAud)
	if err != nil {
		return nil, err
	}

	user, err := s.users.FindByGoogleID(profile.GoogleID)
	if err != nil {
		// belum pernah daftar lewat Google, cek apakah email sudah ada (dulu daftar via password)
		existing, _ := s.users.FindByEmail(profile.Email)
		if existing != nil {
			existing.GoogleID = &profile.GoogleID
			if err := s.users.Update(existing); err != nil {
				return nil, err
			}
			user = existing
		} else {
			user = &domain.User{
				Email:    profile.Email,
				Name:     profile.Name,
				GoogleID: &profile.GoogleID,
			}
			if err := s.users.Create(user); err != nil {
				return nil, err
			}
		}
	}

	return s.issueTokenPair(user.ID, deviceInfo)
}

// RegisterWithPassword: daftar pakai email + password biasa.
func (s *AuthService) RegisterWithPassword(email, password, name string) (*domain.User, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	hashedStr := string(hashed)
	user := &domain.User{
		Email:    email,
		Name:     name,
		Password: &hashedStr,
	}
	if err := s.users.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

// LoginWithPassword: login pakai email + password.
func (s *AuthService) LoginWithPassword(email, password, deviceInfo string) (*TokenPair, error) {
	user, err := s.users.FindByEmail(email)
	if err != nil || user.Password == nil {
		return nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	return s.issueTokenPair(user.ID, deviceInfo)
}

// RefreshAccessToken: tukar refresh token (per device) jadi access token baru.
func (s *AuthService) RefreshAccessToken(rawRefreshToken string) (string, error) {
	hash := authutil.HashToken(rawRefreshToken)
	rt, err := s.refreshTokens.FindByHash(hash)
	if err != nil {
		return "", errors.New("refresh token tidak valid atau sudah expired")
	}
	return authutil.GenerateAccessToken(rt.UserID, s.accessSecret)
}

// Logout: revoke satu refresh token (satu device saja).
func (s *AuthService) Logout(rawRefreshToken string) error {
	hash := authutil.HashToken(rawRefreshToken)
	rt, err := s.refreshTokens.FindByHash(hash)
	if err != nil {
		return nil // sudah tidak valid, anggap sukses
	}
	return s.refreshTokens.Revoke(rt.ID)
}

func (s *AuthService) issueTokenPair(userID uint, deviceInfo string) (*TokenPair, error) {
	access, err := authutil.GenerateAccessToken(userID, s.accessSecret)
	if err != nil {
		return nil, err
	}
	rawRefresh, hash, err := authutil.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}
	if err := s.refreshTokens.Create(&domain.RefreshToken{
		UserID:     userID,
		TokenHash:  hash,
		DeviceInfo: deviceInfo,
		ExpiresAt:  timeNowAdd(authutil.RefreshTokenTTL),
	}); err != nil {
		return nil, err
	}
	return &TokenPair{AccessToken: access, RefreshToken: rawRefresh}, nil
}
