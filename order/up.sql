-- Create a table for orders with an ID, creation time, account ID, and total price.
CREATE TABLE IF NOT EXISTS orders (
    id CHAR(27) PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    account_id CHAR(27) NOT NULL,
    total_price MONEY NOT NULL
);

-- Create a table for order products with an order ID, product ID, quantity, and primary key on the combination of product ID and order ID.
CREATE TABLE IF NOT EXISTS order_products (
    order_id CHAR(27) REFERENCES orders (id) ON DELETE CASCADE,
    product_id CHAR(27),
    quantity INT NOT NULL,
    PRIMARY KEY (product_id, order_id)
);
