package service

import (
	"context"
	"errors"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/umutozd/iam-prototypes/auth"
	"github.com/umutozd/iam-prototypes/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (s *Server) AuthInterceptor(cfg pb.ServiceAuthConfig) func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	methodMap := make(map[string]*pb.AuthMethodConfig, len(cfg.Methods))
	for _, method := range cfg.Methods {
		methodMap[method.FullMethod] = method
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// TODO: parseTokenFromContext doesn't verify signature but it should
		_, ci, err := s.parseTokenFromContext(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "%v", err)
		}

		logrus.Infof("auth: user %s has groups: %v", ci.UserId, ci.Groups)

		if method, ok := methodMap[info.FullMethod]; ok {
			allowed := s.auth.Authorize(ci.UserId, ci.Groups, method.Resource, method.Action)
			if !allowed {
				return nil, status.Error(codes.PermissionDenied, "permission denied")
			}
			return handler(ctx, req)
		}

		// config doesn't contain this method, allow it by default
		return handler(ctx, req)
	}
}

type ClientInfo struct {
	UserId string
	Groups []string
}

func (s *Server) parseTokenFromContext(ctx context.Context) (*auth.JWT, *ClientInfo, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, nil, errors.New("missing metadata")
	}
	au, ok := md["authorization"]
	if !ok || len(au) == 0 {
		au, ok = md["Authorization"]
		if !ok || len(au) == 0 {
			return nil, nil, errors.New("missing authorization header")
		}
	}

	authHeader := au[0]
	if len(authHeader) <= 6 || strings.ToUpper(authHeader[0:6]) != "BEARER" {
		return nil, nil, errors.New("should be a bearer token")
	}
	rawToken := authHeader[7:]
	if len(rawToken) == 0 {
		return nil, nil, errors.New("bearer token is empty")
	}
	if strings.Contains(rawToken, ",") {
		parts := strings.Split(rawToken, ",")
		rawToken = parts[0]
	}

	jwt, err := auth.ParseJWT(rawToken)
	if err != nil {
		return nil, nil, errors.New("invalid token")
	}
	claims, err := jwt.Claims()
	if err != nil {
		return nil, nil, errors.New("invalid token claims")
	}

	ci := &ClientInfo{}
	if sub, ok := claims["sub"]; !ok {
		return nil, nil, errors.New("missing sub claim in token")
	} else if subStr, ok := sub.(string); !ok {
		return nil, nil, errors.New("invalid type for sub claim in token")
	} else {
		ci.UserId = subStr
	}
	if groupsClaim, ok := claims["groups"]; !ok {
		return nil, nil, errors.New("missing groups claim in token")
	} else if groupsSlice, ok := groupsClaim.([]interface{}); !ok {
		return nil, nil, errors.New("invalid type for groups claim in token")
	} else {
		for _, g := range groupsSlice {
			if gs, ok := g.(string); ok {
				ci.Groups = append(ci.Groups, gs)
			} else {
				return nil, nil, errors.New("invalid type for a group in token claims")
			}
		}
	}

	return jwt, ci, nil
}
