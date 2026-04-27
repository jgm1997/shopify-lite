-- name: GetMerchantsProducts :many
select *
from products
where merchant_id = $1;
-- name: GetProduct :one
select *
from products
where id = $1
  and merchant_id = $2;
-- name: CreateProduct :one
insert into products (
    merchant_id,
    "name",
    "description",
    price,
    stock
  )
values ($1, $2, $3, $4, $5)
returning "id",
  merchant_id,
  "name",
  "description",
  price,
  stock,
  created_at;
-- name: UpdateProduct :one
update products
set "name" = $3,
  "description" = $4,
  price = $5,
  stock = $6
where "id" = $1
  and merchant_id = $2
returning "id",
  merchant_id,
  "name",
  "description",
  price,
  stock,
  created_at;
-- name: DeleteProduct :execrows
delete from products
where id = $1
  and merchant_id = $2;
-- name: GetProductByID :one
select *
from products
where "id" = $1;
-- name: GetProductsPaginated :many
select *
from products
limit $1 offset $2;