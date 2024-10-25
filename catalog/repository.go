package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
)

var (
    ErrNotFound = errors.New("entity not found")
)

type Repository interface {
    Close()
    PutProduct(ctx context.Context, p Product) error
    GetProductByID(ctx context.Context, id string) (*Product, error)
    ListProducts(ctx context.Context, skip uint64, take uint64) ([]Product, error)
    ListProductsWithIDs(ctx context.Context, ids []string)([]Product, error)
    SearchProducts(ctx context.Context, query string, skip uint64, take uint64)([]Product, error)
}

type elasticRepository struct {
    client *elasticsearch.TypedClient
}

type productDocument struct {
    Name        string  `json:"name"`
    Description string  `json:"description"`
    Price       float64 `json:"price"`
}

func NewElasticRepository(url string) (Repository, error) {
    elasticCfg := elasticsearch.Config {
        Addresses: []string{url},
    }
    client, err := elasticsearch.NewTypedClient(elasticCfg)
    if err != nil {
        log.Printf("failed to create elasticsearch client: %v", err)
        return nil, err
    }

    exists, err  := client.Indices.Exists("catalog").Do(context.TODO())
    if err != nil {
        log.Println("failed to check index exists: ", err)
        return nil, err
    }
    if !exists {
        res, err := client.Indices.Create(
            "catalog",
        ).Do(context.TODO())
        if err != nil {
            log.Println("failed to create elasticsearch catalog index: ", err)
            return nil, err
        }
        log.Println("create elasticsearch index catalog success: ", res.Index)
    }

    return &elasticRepository{client}, nil
}

func (r *elasticRepository) Close() {}

func (r *elasticRepository) PutProduct(ctx context.Context, p Product) error {
    product_docuemnt := productDocument{
        Name: p.Name,
        Description: p.Description,
        Price: p.Price,
    }

    res, err := r.client.Index("catalog").
        Id(p.ID).
        Request(product_docuemnt).
        Do(ctx)
    if err != nil {
        log.Println("failed to index product: ", err)
        return err
    }

    log.Println("index product document success: ", res.Result)

    return err
}

func (r *elasticRepository) GetProductByID(
    ctx context.Context, id string,
) (*Product, error) {
    res, err := r.client.Get("catalog", id).Do(ctx)
    if err != nil {
        log.Println("failed to get product by id from catalog repository: ", err)
        return nil, err
    }

    p := &productDocument{}
    err = json.Unmarshal(res.Source_, &p)
    if err != nil {
        log.Println("failed to unmarshal productDocument: ", err)
        return nil, err
    }

    return &Product{
        ID: id,
        Name: p.Name,
        Description: p.Description,
        Price: p.Price,
    }, err
}

func (r *elasticRepository) ListProducts(
    ctx context.Context, skip uint64, take uint64,
) ([]Product, error) {
    res, err := r.client.Search().
        Index("catalog").
        Request(&search.Request{
            Query: &types.Query{MatchAll: &types.MatchAllQuery{}},
        }).
        Do(context.TODO())
    
    if err != nil {
        log.Println("failed to list all products from catalog repository: ", err)
        return nil, err
    }

    products := []Product{}
    for _, hit := range res.Hits.Hits {
        p := productDocument{}
        if err = json.Unmarshal(*&hit.Source_, &p); err == nil {
            products = append(products, Product{
                ID: *hit.Id_,
                Name: p.Name,
                Description: p.Description,
                Price: p.Price,
            })
        }
    }

    return products, err
}

func (r *elasticRepository) ListProductsWithIDs(
    ctx context.Context, ids []string,
)([]Product, error) {

    res, err := r.client.Search().
        Index("catalog").
        Request(&search.Request{
            Query: &types.Query{
                Ids: &types.IdsQuery{
                    Values: ids,
                },
            },
        }).Do(ctx)

    if err != nil {
        log.Println("failed to list products with ids from catalog repository: ", err)
        return nil, err
    }

    products := []Product{}
    for _, hit := range res.Hits.Hits {
        p := productDocument{}
        if err = json.Unmarshal(*&hit.Source_, &p); err == nil {
            products = append(products, Product{
                ID: *hit.Id_,
                Name: p.Name,
                Description: p.Description,
                Price: p.Price,
            })
        }
    }
    log.Println("catalog: repository: products: ", products)

    return products, err
}

func (r *elasticRepository) SearchProducts(
    ctx context.Context, query string, skip uint64, take uint64,
)([]Product, error) {
    skip_int := int(skip)
    take_int := int(take)
    res, err := r.client.Search().
        Index("catalog").
        Request(&search.Request{
            Query: &types.Query{
                MultiMatch: &types.MultiMatchQuery{
                    Fields: []string{"name", "description"},
                    Query: query,
                },
            },
            From: &skip_int,
            Size: &take_int,
        }).Do(ctx)
    if err != nil {
        log.Println("failed to search products with query from catalog repository: ", err)
        return nil, err
    }

    products := []Product{}
    for _, hit := range res.Hits.Hits {
        p := productDocument{}
        if err = json.Unmarshal(*&hit.Source_, &p); err == nil {
            products = append(products, Product{
                ID: *hit.Id_,
                Name: p.Name,
                Description: p.Description,
                Price: p.Price,
            })
        }
    }

    return products, err
}
