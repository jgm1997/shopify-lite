-- name: GetMerchantDashboard :one
select count(*) as total_products,
   coalesce(sum(stock), 0) as total_stock,
   coalesce(
      sum(price * stock),
      0
   ) as inventory_value,
   count(*) filter(
      where stock = 0
   ) as out_of_stock
from products
where merchant_id = $1;
-- name: CreateMerchant :one
insert into merchants (user_id, store_name, "description")
values ($1, $2, $3)
returning "id",
   user_id,
   store_name,
   "description",
   created_at;
-- name: GetMerchantByUserID :one
select "id",
   user_id,
   store_name,
   "description",
   created_at
from merchants
where user_id = $1;