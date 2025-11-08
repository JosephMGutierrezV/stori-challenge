CREATE TABLE IF NOT EXISTS transactions.transactions
(
    id         bigserial PRIMARY KEY,
    bucket     varchar(255),
    object_key varchar(512),
    date     timestamptz,
    amount     numeric,
    created_at timestamptz DEFAULT now()
);