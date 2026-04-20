create table merchants (
   id         serial primary key,
   user_id    int unique not null
      references users ( id )
         on delete cascade,
   store_name text not null,
   description text,
   created_at timestamptz not null default now()
);