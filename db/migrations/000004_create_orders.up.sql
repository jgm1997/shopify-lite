create type order_status as enum (
    'pending', 'confirmed', 'shipped', 'delivered', 'cancelled'
);

create table orders (
    id serial primary key,
    customer_id int not null references users(id),
    status order_status not null default 'pending',
    total numeric(10,2) not null,
    created_at timestamptz not null default now()
);

create table order_items (
    id serial primary key,
    order_id int not null references orders(id) on delete cascade,
    product_id int not null references products(id),
    quantity int not null check (quantity > 0),
    unit_price numeric(10,2) not null
);