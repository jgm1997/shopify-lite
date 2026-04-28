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
where id = $1
    and customer_id = $2;
-- name: GetOrderByIDForMerchant :one
select o.*
from orders o
    join order_items oi on oi.order_id = o.id
    join products p on p.id = oi.product_id
where o.id = $1
    and p.merchant_id = $2
limit 1;
-- name: UpdateOrderStatus :one
update orders
set status = $2
where id = $1
returning *;

-- name: GetOrderItemsForMerchant :many
select oi.*
from order_items oi
    join products p on p.id = oi.product_id
where oi.order_id = $1
    and p.merchant_id = $2;