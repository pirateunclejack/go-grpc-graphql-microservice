package order

import (
	"context"
	"log"
	"time"

	"github.com/pirateunclejack/go-grpc-graphql-microservice/order/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
    conn *grpc.ClientConn
    service pb.OrderServiceClient
}

func NewClient(url string) (*Client, error){
    opts := grpc.WithTransportCredentials(insecure.NewCredentials())
    conn, err := grpc.NewClient(url, opts)
    if err != nil {
        log.Println("failed to create new grpc client from order client: ", err)
        return nil, err
    }

    c := pb.NewOrderServiceClient(conn)
    return &Client{conn: conn, service: c}, nil
}

func (c *Client) Close() {
    c.conn.Close()
}


func (c *Client) PostOrder(
    ctx context.Context, accountID string, products []OrderedProduct,
) (*Order, error){
    protoProducts := []*pb.PostOrderRequest_OrderProduct{}
    for _, p := range products {
        protoProducts = append(protoProducts, &pb.PostOrderRequest_OrderProduct{
            ProductId: p.ID,
            Quantity: p.Quantity,
        })
    }

    r, err := c.service.PostOrder(
        ctx,
        &pb.PostOrderRequest{
            AccountId: accountID,
            Products: protoProducts,
        },
    )
    if err != nil {
        log.Println("failed to post order from order client: ", err)
        return nil, err
    }

    newOrder := r.Order
    newOrderCreatedAt := time.Time{}
    newOrderCreatedAt.UnmarshalBinary(newOrder.CreatedAt)
    return &Order{
        ID: newOrder.Id,
        CreatedAt: newOrderCreatedAt,
        TotalPrice: newOrder.TotalPrice,
        Products: products,
    }, nil

}

func (c *Client) GetOrdersForAccount(
    ctx context.Context, accountID string,
) ([]Order, error){
    r, err := c.service.GetOrdersForAccount(
        ctx,
        &pb.GetOrdersForAccountRequest{
            AccountId: accountID,
        },
    )
    if err != nil {
        log.Println("failed to get orders for account from order client: ", err)
        return nil, err
    }
    orders := []Order{}
    for _, orderProto := range r.Orders {
        newOrder := Order{
            ID: orderProto.Id,
            TotalPrice: orderProto.TotalPrice,
            AccountID: orderProto.AccountId,
        }
        newOrder.CreatedAt = time.Time{}
        newOrder.CreatedAt.UnmarshalBinary(orderProto.CreatedAt)
        products := []OrderedProduct{}
        for _, p := range orderProto.Products {
            products = append(products, OrderedProduct{
                ID: p.Id,
                Quantity: p.Quantity,
                Name: p.Name,
                Description: p.Description,
                Price: p.Price,
            })
        }
        newOrder.Products = products
        orders = append(orders, newOrder)
        log.Println("orders from client: ", orders)
    }

    return orders, nil
}
