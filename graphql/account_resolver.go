package main

import (
	"context"
	"log"
	"time"
)

type accountResolver struct {
    server *Server
}

func (r *accountResolver) Orders(
    ctx context.Context, obj *Account,
) ([]*Order, error){
    ctx, cancel := context.WithTimeout(ctx, 3 * time.Second)
    defer cancel()

    orderList, err := r.server.orderClient.GetOrdersForAccount(
        ctx, obj.ID,
    )
    if err != nil {
        log.Println(err)
        return nil, err
    }

    var orders []*Order

    for _, o := range orderList {
        var products []*OrderedProduct
        for _, p := range o.Products {
            products = append(products, &OrderedProduct{
                ID:             p.ID,
                Name:           p.Name,
                Price:          p.Price,
                Quantity:       int(p.Quantity),
                Description:    p.Description,
            })

        }
        orders = append(orders, &Order{
            ID: o.ID,
            CreatedAt: o.CreatedAt,
            TotalPrice: float64(o.TotalPrice),
            Products: products,
        })
    }

    return orders, nil
}
