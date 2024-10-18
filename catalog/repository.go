package catalog

import (
	"context"
	"encoding/json"
	"errors"

	// "github.com/elastic/go-elasticsearch"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/update"
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
    Price       float32 `json:"price"`
}

func NewElasticRepository(url string) (Repository, error) {
    elasticCfg := elasticsearch.Config {
        Addresses: []string{url},
    }
    client, err := elasticsearch.NewTypedClient(elasticCfg)
    if err != nil {
        return nil, err
    }

    return &elasticRepository{client}, nil
}

func (r *elasticRepository) Close() {
    
}

func (r *elasticRepository) PutProduct(ctx context.Context, p Product) error {
    data, err := json.Marshal(productDocument{
        Name: p.Name,
        Description: p.Description,
        Price: p.Price,
    })
    if err != nil {
        return err
    }

    _, err = r.client.Update("catalog", p.ID).Request(
        &update.Request{
            Doc: data,
        },
    ).Do(ctx)
    if err != nil {
        return err
    }
    ctx.Done()
    return err
}

func (r *elasticRepository) GetProductByID(
    ctx context.Context, id string,
) (*Product, error) {
    res, err := r.client.Get("catalog", id).Do(ctx)
    if err != nil {
        return nil, err
    }

    p := &productDocument{}
    p_byte, err := res.Source_.MarshalJSON()
    if err != nil {
        return nil, err
    }
    err = json.Unmarshal(p_byte, &p)
    if err != nil {
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
        return nil, err
    }

    products := []Product{}
    for _, hit := range res.Hits.Hits {
        p := productDocument{}
        if err = json.Unmarshal(*&hit.Source_, &p); err == nil {
            products = append(products, Product{
                ID: hit.Id_,
                Name: p.Name,
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
        return nil, err
    }

    products := []Product{}
    for _, hit := range res.Hits.Hits {
        p := productDocument{}
        if err = json.Unmarshal(*&hit.Source_, &p); err == nil {
            products = append(products, Product{
                ID: hit.Id_,
                Name: p.Name,
            })
        }
    }

    return products, err
}

func (r *elasticRepository) SearchProducts(
    ctx context.Context, query string, skip uint64, take uint64,
)([]Product, error) {
    skip_int := int(skip)
    take_int := int(take)
    // skip_int, err := strconv.Atoi(skip)
    // if err != nil {
    //     return nil, err
    // }
    // take_int, err := strconv.Atoi(take)
    // if err != nil {
    //     return nil, err
    // }
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
        return nil, err
    }
    
    products := []Product{}
    for _, hit := range res.Hits.Hits {
        p := productDocument{}
        if err = json.Unmarshal(*&hit.Source_, &p); err == nil {
            products = append(products, Product{
                ID: hit.Id_,
                Name: p.Name,
            })
        }
    }

    return products, err
}
