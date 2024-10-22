package main

import (
	"context"
	"log"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/pirateunclejack/go-grpc-graphql-microservice/catalog"
	"github.com/sethvargo/go-retry"
)

type Config struct {
    DatabaseURL string `envconfig:"DATABASE_URL"`
}

func main() {
    var cfg Config
    err := envconfig.Process("", &cfg)
    if err!= nil {
        log.Fatal(err)
    }

    var r catalog.Repository

    ctx := context.Background()
    if err := retry.Constant(ctx, time.Second * 1 , func(ctx context.Context) error {
        r, err = catalog.NewElasticRepository(cfg.DatabaseURL)
        if err != nil {
            log.Println(err)
            return retry.RetryableError(err)
        }
        return nil
    }); err != nil {
        log.Fatal(err)
    }

    defer r.Close()
    log.Println("Listening on port 8080...")

    s := catalog.NewService(r)
    log.Fatal(catalog.ListenGRPC(s, 8080))
}
