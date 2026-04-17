create type user_role as enum('merchant', 'customer');
create table users (
    id serial primary key,
    email text unique not null,
    password_hash text not null,
    role user_role not null,
    created_at timestamptz not null default now()
);