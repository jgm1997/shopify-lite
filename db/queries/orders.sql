-- name: CreateOrder :one
insert into orders (customer_id, status, total)
values ($1, 'pending', $2)
returning *;
-- name: CreateOrderItem :one
insert into order_items (order_id, product_id, quantity, unit_price)
values ($1, $2, $3, $4)
returning *;
-- name: GetProductForUpdate :one
select id,
    merchant_id,
    stock,
    price
from products
where id = $1 for
update;
-- name: DeductStock :one
update products
set stock = stock - $2
where id = $1
    and stock >= $2
returning stock;
-- name: GetCustomerOrders :many
select o.*,
    oi.product_id,
    oi.quantity,
    oi.unit_price
from orders o
    join order_items oi on oi.order_id = o.id
where o.customer_id = $1
order by o.created_at desc;
-- name: GetOrderByID :one
select *
from orders
where id = $1 and customer_id = $2;