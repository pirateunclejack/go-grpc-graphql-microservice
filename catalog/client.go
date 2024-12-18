package catalog

import (
	"context"
	"log"

	"github.com/pirateunclejack/go-grpc-graphql-microservice/catalog/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
    conn    *grpc.ClientConn
    service pb.CatalogServiceClient
}

func NewClient(url string) (*Client, error) {
    opts := grpc.WithTransportCredentials(insecure.NewCredentials())
    conn, err := grpc.NewClient(url, opts)
    if err != nil {
        log.Println("failed to create catalog grpc client: ", err)
        return nil, err
    }

    c := pb.NewCatalogServiceClient(conn)
    return &Client{conn: conn, service: c}, nil
}

func (c *Client) Close() {
    c.conn.Close()
}


func (c *Client) PostProduct(
    ctx context.Context, name, description string, price float64,
) (*Product, error) {
    r, err := c.service.PostProduct(
        ctx, &pb.PostProductRequest{
            Name: name,
            Description: description,
            Price: price,
        },
    )
    if err != nil {
        log.Println("failed to post product from catalog client: ", err)
        return nil, err
    }

    return &Product{
        ID: r.Product.Id,
        Name: r.Product.Name,
        Description: r.Product.Description,
        Price: r.Product.Price,
    }, nil
}

func (c *Client) GetProduct(ctx context.Context, id string) (*Product, error) {
    r, err := c.service.GetProduct(
        ctx, 
        &pb.GetProductRequest{
            Id: id,
        },
    )
    if err != nil {
        log.Println("failed to get product from catalog client: ", err)
        return nil, err
    }

    return &Product{
        ID: r.Product.Id,
        Name: r.Product.Name,
        Description: r.Product.Description,
        Price: r.Product.Price,
    }, nil
}

func (c *Client) GetProducts(
    ctx context.Context,
    skip, take uint64,
    ids []string, query string,
) (*[]Product, error) {
    r, err := c.service.GetProducts(
        ctx,
        &pb.GetProductsRequest{
            Skip: skip,
            Take: take,
            Ids: ids,
            Query: query,
        },
    )
    if err != nil {
        log.Println("failed to get products from catalog client: ", err)
        return nil, err
    }

    products := []Product{}
    for _, p := range r.Products{
        products = append(products, Product{
            ID: p.Id,
            Name: p.Name,
            Description: p.Description,
            Price: p.Price,
        })
    }
    log.Println("catalog: client: products: ", products)

    return &products, err
}
