create table products (
   id          serial primary key,
   merchant_id int not null
      references merchants ( id )
         on delete cascade,
   name        text not null,
   description text,
   price       numeric(10,2) not null,
   stock       int not null default 0,
   created_at  timestamptz not null default now()
);