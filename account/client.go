package account

import (
	"context"
	"log"

	"github.com/pirateunclejack/go-grpc-graphql-microservice/account/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
    conn    *grpc.ClientConn
    service pb.AccountServiceClient
}

func NewClient(url string) (*Client, error) {
    opts := grpc.WithTransportCredentials(insecure.NewCredentials())
    conn, err := grpc.NewClient(url, opts)
    if err != nil {
        log.Println("failed to create account grpc client: ", err)
        return nil, err
    }

    c := pb.NewAccountServiceClient(conn)
    return &Client{conn: conn, service: c}, nil
}

func (c *Client) Close() {
    c.conn.Close()
}

func (c *Client) PostAccount(ctx context.Context, name string) (*Account, error) {
    r, err := c.service.PostAccount(
        ctx,
        &pb.PostAccountRequest{Name: name},
    )
    if err != nil {
        log.Println("failed to post account from account client: ", err)
        return nil, err
    }

    return &Account{
        ID: r.Account.Id,
        Name: r.Account.Name,
    }, nil
}

func (c *Client) GetAccount(ctx context.Context, id string) (*Account, error) {
    r, err := c.service.GetAccount(
        ctx,
        &pb.GetAccountRequest{Id: id},
    )
    if err != nil {
        log.Println("failed to get account from account client: ", err)
        return nil, err
    }
    return &Account{
        ID: r.Account.Id,
        Name: r.Account.Name,
    }, nil
}

func (c *Client) GetAccounts(ctx context.Context, skip uint64, take uint64) ([]Account, error){
    r, err :=c.service.GetAccounts(
        ctx,
        &pb.GetAccountsRequest{Skip:skip, Take:take},
    )
    if err != nil {
        log.Println("failed to get accounts from account client: ", err)
        return nil, err
    }

    accounts := []Account{}
    for _, account := range r.Accounts {
        accounts = append(
            accounts, 
            Account{
                ID: account.Id,
                Name: account.Name,
            },
        )
    }
    return accounts, nil
}
