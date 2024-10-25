package main

import (
	"log"
	"net/http"

	// "github.com/99designs/gqlgen/graphql/handler"
	// "github.com/99designs/gqlgen/graphql/playground"
	"github.com/99designs/gqlgen/handler"
	"github.com/kelseyhightower/envconfig"
)

type AppConfig struct {
    AccountURL string `envconfig:"ACCOUNT_SERVICE_URL"`
    CatalogURL string `envconfig:"CATALOG_SERVICE_URL"`
    OrderURL   string `envconfig:"ORDER_SERVICE_URL"`
}

func main() {
    var cfg AppConfig
    err := envconfig.Process("", &cfg)
    if err != nil {
        log.Fatal("failed to get graphql config with envconfig: ", err)
    }

    s, err := NewGraphQLServer(
        cfg.AccountURL, cfg.CatalogURL, cfg.OrderURL,
    )
    if err != nil {
        log.Fatal("failed to create graphql server from graphql: ", err)
    }

    http.Handle(
        "/graphql",
        // handler.New(s.ToExecutableSchema()),
        handler.GraphQL(s.ToExecutableSchema()),
    )

    http.Handle(
        "/playground",
        handler.Playground("jack", "/graphql"),
        // playground.Handler("jack", "/graphql"),
    )

    log.Fatal(http.ListenAndServe(":8080", nil))
}
