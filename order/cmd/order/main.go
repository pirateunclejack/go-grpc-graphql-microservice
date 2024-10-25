package main

import (
	"context"
	"log"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/pirateunclejack/go-grpc-graphql-microservice/order"
	"github.com/sethvargo/go-retry"
)

type Config struct {
    DatabaseURL string `envconfig:"DATABASE_URL"`
    AccountURL  string `envconfig:"ACCOUNT_SERVICE_URL"`
    CatalogURL  string `envconfig:"CATALOG_SERVICE_URL"`
}

func main() {
    var cfg Config
    err := envconfig.Process("", &cfg)
    if err != nil {
        log.Println("failed to get order config with envconfig: ", err)
        log.Fatal(err)
    }

    var r order.Repository

    ctx := context.Background()
    if err := retry.Constant(ctx, time.Second * 1 , func(ctx context.Context) error {
        r, err = order.NewPostgresRepository(cfg.DatabaseURL)
        if err != nil {
            log.Println("failed to create order postgres repository: ", err)
            return retry.RetryableError(err)
        }
        return nil
    }); err != nil {
        log.Fatal("failed to retry create order postgres repository: ", err)
    }

    defer r.Close()
    log.Println("Listening on port 8080...")

    s := order.NewService(r)
    log.Fatal(order.ListenGRPC(s, cfg.AccountURL, cfg.CatalogURL, 8080))
}
