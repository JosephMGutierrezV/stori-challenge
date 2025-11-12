CREATE TABLE IF NOT EXISTS transactions.account_summaries
(
    id            bigserial
        primary key,
    bucket        varchar(255),
    object_key    varchar(512),
    total_balance text,
    raw_summary   text,
    created_at    timestamp with time zone
);