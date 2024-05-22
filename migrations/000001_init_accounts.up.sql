create table if not exists accounts(
    id bigserial not null primary key,
    name varchar(50) not null,
    balance numeric(12,2) not null default 0,
    created_at timestamp not null default now()
)