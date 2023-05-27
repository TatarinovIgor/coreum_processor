create table if not exists assets
(
    id             bigserial
    primary key,
    created_at     timestamp with time zone not null,
    updated_at     timestamp with time zone not null,
    deleted_at     timestamp with time zone,
    blockchain     varchar(32)              not null,
    code           varchar(32)              not null,
    issuer         varchar(128)             not null,
    name           varchar(32)              not null,
    description    varchar                  not null default '',
    merchant_owner bigint                            default null,
    constraint fk_asset_merchant_list
    foreign key (merchant_owner)
    references merchant_list (id)
    )