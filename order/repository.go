package order

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/lib/pq"
)

type Repository interface {
    Close()
    PutOrder(ctx context.Context, o Order) error
    GetOrdersForAccount(ctx context.Context, accountID string) ([]Order, error)
}

type postgresRepository struct {
    db *sql.DB
}

func NewPostgresRepository(url string) (*postgresRepository, error){
    db, err := sql.Open("postgres", url)
    if err != nil {
        log.Println("failed to create postgres order repository from order repository: ", err)
        return nil, err
    }

    err = db.Ping()
    if err != nil {
        log.Println("failed to connect to postgres order repository: ", err)
        return nil, err
    }

    return &postgresRepository{db: db}, nil
}

func (r *postgresRepository) Close(){
    r.db.Close()
}

func (r *postgresRepository) PutOrder(ctx context.Context, o Order) error {

    log.Println("order: repository: order: ", o)
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf(
            "failed to start put order transaction from order repository: %w",
            err,
        )
    }

    defer func() {
        if err != nil {
            log.Println("failed to put order, rollback, : ", err)
            tx.Rollback()
            return
        }
        err = tx.Commit()
    }()

    _, err = tx.ExecContext(
        ctx,
        "INSERT INTO orders (id, created_at, account_id, total_price) VALUES ($1,$2,$3,$4)",
        o.ID,
        o.CreatedAt,
        o.AccountID,
        o.TotalPrice,
    )
    if err != nil {
        log.Println("failed to insert order from order repository: ", err)
        return fmt.Errorf("failed to insert order from order repository: %w", err)
    }

    stmt, _ := tx.PrepareContext(ctx, pq.CopyIn(
        "order_products",
        "order_id",
        "product_id",
        "quantity",
    ))
    for _, p := range o.Products{
        _, err = stmt.ExecContext(ctx, o.ID, p.ID, p.Quantity)
        if err != nil {
        log.Println("failed to insert order product from order repository: ", err)
        return fmt.Errorf("failed to insert order product from order repository: %w", err)
        }
    }

    _, err = stmt.ExecContext(ctx)
    if err != nil {
        log.Println("failed to commit order products from order repository: ", err)
        return fmt.Errorf("failed to commit order products from order repository: %w", err)
    }

    stmt.Close()
    return err
}

func (r *postgresRepository) GetOrdersForAccount(
    ctx context.Context, accountID string,
) ([]Order, error){
    rows, err := r.db.QueryContext(
        ctx,
        `SELECT
        o.id,
        o.created_at,
        o.account_id,
        o.total_price::money::numeric::float8,
        op.product_id,
        op.quantity
        FROM orders o JOIN order_products op ON (o.id = op.order_id)
        WHERE o.account_id=$1
        ORDER BY o.id`,
        accountID,
    )
    if err != nil {
        log.Println("failed to get orders from order repository: ", err)
        return nil, fmt.Errorf("failed to get orders from order repository: %w", err)
    }
    defer rows.Close()

    orders := []Order{}
    order := &Order{}
    lastOrder := &Order{}
    orderedProduct := &OrderedProduct{}
    products := []OrderedProduct{}
    newOrder := Order{}

    // Scan rows into Order structs
    for rows.Next() {
        if err = rows.Scan(
            &order.ID,
            &order.CreatedAt,
            &order.AccountID,
            &order.TotalPrice,
            &orderedProduct.ID,
            &orderedProduct.Quantity,
        ); err != nil {
            return nil, err
        }

        log.Println("row: ", order.ID, order.CreatedAt, order.AccountID, order.TotalPrice, orderedProduct.ID, orderedProduct.Quantity)
        log.Println("lastorder: ", lastOrder)
        // Scan order

        if lastOrder.ID == "" {
            products = append(products, OrderedProduct{
                ID:       orderedProduct.ID,
                Quantity: orderedProduct.Quantity,
            })
            newOrder = Order{
                ID:         order.ID,
                AccountID:  order.AccountID,
                CreatedAt:  order.CreatedAt,
                TotalPrice: order.TotalPrice,
                Products:   products,
            }
        } else {
            if lastOrder.ID == order.ID {
                // newOrder.Products = append(newOrder.Products, OrderedProduct{
                //     ID:       orderedProduct.ID,
                //     Quantity: orderedProduct.Quantity,
                // })
                products = append(products, OrderedProduct{
                    ID:       orderedProduct.ID,
                    Quantity: orderedProduct.Quantity,
                })
                newOrder = Order{
                    ID:         order.ID,
                    AccountID:  order.AccountID,
                    CreatedAt:  order.CreatedAt,
                    TotalPrice: order.TotalPrice,
                    Products:   products,
                }
            } else {
                orders = append(orders, newOrder)
                products = []OrderedProduct{}
                products = append(products, OrderedProduct{
                    ID:       orderedProduct.ID,
                    Quantity: orderedProduct.Quantity,
                })
                newOrder = Order{
                    ID:         order.ID,
                    AccountID:  order.AccountID,
                    CreatedAt:  order.CreatedAt,
                    TotalPrice: order.TotalPrice,
                    Products:   products,
                }
            }
        }

        *lastOrder = *order

        log.Println("orders: ", orders)

    }
    orders = append(orders, newOrder)

    log.Println("final orders: ", orders)

    // orders := []Order{}
	// order := &Order{}
	// lastOrder := &Order{}
	// orderedProduct := &OrderedProduct{}
	// products := []OrderedProduct{}

	// // Scan rows into Order structs
	// for rows.Next() {
	// 	if err = rows.Scan(
	// 		&order.ID,
	// 		&order.CreatedAt,
	// 		&order.AccountID,
	// 		&order.TotalPrice,
	// 		&orderedProduct.ID,
	// 		&orderedProduct.Quantity,
	// 	); err != nil {
	// 		return nil, err
	// 	}

    //     log.Println("row: ", order.ID, order.CreatedAt, order.AccountID, order.TotalPrice, orderedProduct.ID, orderedProduct.Quantity)
    //     log.Println("lastorder: ", lastOrder)

	// 	// Scan order
	// 	if lastOrder.ID != "" && lastOrder.ID != order.ID {
	// 		newOrder := Order{
	// 			ID:         lastOrder.ID,
	// 			AccountID:  lastOrder.AccountID,
	// 			CreatedAt:  lastOrder.CreatedAt,
	// 			TotalPrice: lastOrder.TotalPrice,
	// 			Products:   products,
	// 		}
	// 		orders = append(orders, newOrder)
	// 		products = []OrderedProduct{}
	// 	}
	// 	// Scan products
	// 	products = append(products, OrderedProduct{
	// 		ID:       orderedProduct.ID,
	// 		Quantity: orderedProduct.Quantity,
	// 	})

	// 	*lastOrder = *order
    //     log.Println("orders: ", orders)
	// }

	// // Add last order (or first :D)
	// if lastOrder != nil {
	// 	newOrder := Order{
	// 		ID:         lastOrder.ID,
	// 		AccountID:  lastOrder.AccountID,
	// 		CreatedAt:  lastOrder.CreatedAt,
	// 		TotalPrice: lastOrder.TotalPrice,
	// 		Products:   products,
	// 	}
	// 	orders = append(orders, newOrder)
	// }

    if err := rows.Err(); err != nil {
        log.Println("failed to get orders from order repository: ", err)
        return nil, fmt.Errorf("failed to get orders from order repository: %w", err)
    }

    log.Println("final orders: ", orders)
    return orders, err
}
