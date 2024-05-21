CREATE TYPE operation AS ENUM('withdrawal', 'deposit', 'transfer');

create table if not exists transactions(
    id bigserial not null primary key,
    account1 bigint not null references accounts(id),
    account2 bigint references accounts(id),
    amount numeric(12,2) not null,
    operation operation null,
    created_at timestamp not null default now()
)