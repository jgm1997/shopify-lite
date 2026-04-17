-- name: CreateUser :one
insert into users(email, password_hash, "role") values ($1, $2, $3) returning "id", email, password_hash, "role", created_at;

-- name: GetUserByEmail :one
select "id", email, password_hash, "role", created_at from users where email = $1;

-- name: GetUserByID :one
select "id", email, password_hash, "role", created_at from users where id = $1;