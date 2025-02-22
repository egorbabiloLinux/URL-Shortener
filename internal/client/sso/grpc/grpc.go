package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	ssov1 "github.com/egorbabiloLinux/protos/gen/go/sso"
	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
    api ssov1.AuthClient
}

func New(
    ctx context.Context,
    log *slog.Logger,
    addr string,
    timeout time.Duration,
    retriesCount int,
) (*Client, error) {
    const op = "grpc.New"

    retryOpts := []grpcretry.CallOption{
        grpcretry.WithCodes(codes.NotFound, codes.Aborted, codes.DeadlineExceeded),
    }

    logOpts := []grpclog.Option{
        grpclog.WithLogOnEvents(grpclog.PayloadReceived, grpclog.PayloadSent),
    }
    
    cc, err := grpc.NewClient(
    addr, 
    grpc.WithTransportCredentials(insecure.NewCredentials()),
    grpc.WithChainUnaryInterceptor(
        grpclog.UnaryClientInterceptor(InterceptorLogger(log), logOpts...),
        grpcretry.UnaryClientInterceptor(retryOpts...),
    ))
    if err != nil {
        return nil, fmt.Errorf("%s: %w", op, err)
    }

    grpcClient := ssov1.NewAuthClient(cc)

    return &Client{
        api: grpcClient,
    }, nil
}

func (c *Client) IsAdmin(ctx context.Context, userID int64) (bool, error) {
    const op = "grpc.IsAdmin"

    resp, err := c.api.IsAdmin(ctx, &ssov1.IsAdminRequest{
        UserId: userID,
    })
    if err != nil {
        return false, fmt.Errorf("%s: %w", op, err)
    }

    return resp.IsAdmin, nil
}

func (c *Client) Register(
    ctx context.Context, 
    email string, 
    password string,
) (int64, error) {
    const op = "grpc.Register"

    resp, err := c.api.Register(ctx, &ssov1.RegisterRequest{
        Email: email,
        Password: password,
    })
    if err != nil {
        return 0, fmt.Errorf("%s: %w", op, err)
    }

    return resp.UserId, nil
}

func (c *Client) Login (
    ctx context.Context,
    email string, 
	password string,
	appID int32,
) (string, error) {
    const op = "grpc.Login"

    resp, err := c.api.Login(ctx, &ssov1.LoginRequest{
        Email: email,
        Password: password,
        AppId: appID,
    })
    if err != nil {
        return "", fmt.Errorf("%s: %w", op, err)
    }

    return resp.Token, nil
}

func InterceptorLogger(l *slog.Logger) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, lvl grpclog.Level, msg string, fields ...any){
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}