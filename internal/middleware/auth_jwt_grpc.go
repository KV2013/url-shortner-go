package middleware

import (
	"context"
	"strings"

	"github.com/KV2013/url-shortner-go/internal/config"
	"github.com/KV2013/url-shortner-go/internal/service/auth"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const authorizationHeader = "authorization"
const bearerPrefix = "Bearer "

func AuthJWTUnaryInterceptor(cfg *config.Config, logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		var userID string

		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			for _, h := range md.Get(authorizationHeader) {
				if token, found := strings.CutPrefix(h, bearerPrefix); found {
					id, parseErr := auth.GetUserID(token, cfg.JWTSecretKey)
					if parseErr == nil && id != "" {
						userID = id
						break
					}
				}
			}
		}

		if userID == "" {
			userID = uuid.New().String()
			if token, tokenErr := auth.GenerateAccessToken(userID, cfg.JWTSecretKey); tokenErr == nil {
				if setErr := grpc.SetHeader(ctx, metadata.Pairs(authorizationHeader, bearerPrefix+token)); setErr != nil {
					logger.Warn("gRPC AuthJWT: не удалось установить authorization header", zap.Error(setErr))
				}
			} else {
				logger.Warn("gRPC AuthJWT: не удалось сгенерировать токен", zap.Error(tokenErr))
			}
		}

		ctx = context.WithValue(ctx, UserIDContextKey, userID)
		return handler(ctx, req)
	}
}
