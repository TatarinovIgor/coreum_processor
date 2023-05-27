create table merchants
(
    id         bigserial primary key unique,
    created_at timestamp with time zone not null,
    updated_at timestamp with time zone not null,
    deleted_at timestamp with time zone default null,
    key        varchar(64)              not null unique,
    value      text                     not null,
    ttl        bigint
);