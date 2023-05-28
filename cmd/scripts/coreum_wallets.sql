create table coreum_wallets
(
    id          bigserial primary key,
    created_at  timestamp with time zone not null,
    updated_at  timestamp with time zone not null,
    deleted_at  timestamp with time zone,
    merchant_id varchar(64)              not null,
    external_id varchar(64)              not null,
    key         varchar(128)             not null unique,
    value       text                     not null
);

create unique index merch_user_idx
    on coreum_wallets (merchant_id, external_id);

