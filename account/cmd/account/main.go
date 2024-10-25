package main

import (
	"context"
	"log"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/pirateunclejack/go-grpc-graphql-microservice/account"
	"github.com/sethvargo/go-retry"
)

type Config struct {
    DatabaseURL string `envconfig:"DATABASE_URL"`
}

func main() {
    var cfg Config
    err := envconfig.Process("", &cfg)
    if err != nil {
        log.Fatal("failed to get account config with envconfig: ", err)
    }

    var r account.Repository

    ctx := context.Background()
    if err := retry.Constant(ctx, time.Second * 1 , func(ctx context.Context) error {
        r, err = account.NewPostgresRepository((cfg.DatabaseURL))
        if err != nil {
            log.Println("failed to create account postgres repository: ", err)
            return retry.RetryableError(err)
        }
        return nil
    }); err != nil {
        log.Fatal("failed to retry to create account postgres repository: ", err)
    }

    defer r.Close()
    log.Println("Listening on port 8080...")

    s := account.NewService(r)
    log.Fatal(account.ListenGRPC(s, 8080))
}
