package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"

	"korp/internal/mailer"
)

const tokenVerificacaoValidade = 24 * time.Hour

type Service struct {
	repo      *Repository
	mail      *mailer.Mailer
	jwtSecret string
	baseURL   string
}

func NewService(repo *Repository, mail *mailer.Mailer, jwtSecret, baseURL string) *Service {
	return &Service{repo: repo, mail: mail, jwtSecret: jwtSecret, baseURL: baseURL}
}

func (s *Service) Cadastrar(ctx context.Context, email, senha string) (User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}

	token := gerarTokenAleatorio()
	expiraEm := time.Now().Add(tokenVerificacaoValidade)

	user, err := s.repo.Criar(ctx, email, string(hash), token, expiraEm)
	if err != nil {
		return User{}, err
	}

	link := fmt.Sprintf("%s/auth/verificar?token=%s", s.baseURL, token)
	corpo := fmt.Sprintf("Confirme seu cadastro clicando no link: %s", link)
	if err := s.mail.Enviar(user.Email, "Confirme seu cadastro", corpo); err != nil {
		slog.Error("falha ao enviar email de verificacao", "error", err, "email", user.Email)
	}

	return user, nil
}

func (s *Service) VerificarEmail(ctx context.Context, token string) error {
	user, err := s.repo.BuscarPorToken(ctx, token)
	if err != nil {
		return err
	}
	if user.TokenVerificacaoExpiraEm == nil || time.Now().After(*user.TokenVerificacaoExpiraEm) {
		return ErrTokenExpirado
	}
	return s.repo.MarcarVerificado(ctx, user.ID)
}

func (s *Service) Login(ctx context.Context, email, senha string) (string, error) {
	user, err := s.repo.BuscarPorEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.SenhaHash), []byte(senha)); err != nil {
		return "", ErrCredenciaisInvalidas
	}
	if !user.Verificado {
		return "", ErrEmailNaoVerificado
	}
	return gerarTokenJWT(user.ID, user.Email, s.jwtSecret)
}

func gerarTokenAleatorio() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
