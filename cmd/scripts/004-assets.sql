create table if not exists assets
(
    id             bigserial primary key,
    created_at     timestamp with time zone not null,
    updated_at     timestamp with time zone not null,
    deleted_at     timestamp with time zone,
    blockchain     varchar(32)              not null,
    code           varchar(32)              not null,
    issuer         varchar(128)                      default null,
    name           varchar(32)              not null,
    description    varchar                  not null default '',
    merchant_owner bigint                            default null,
    status         varchar                  not null default 'pending',
    type           varchar                  not null default 'fungible',
    features       json                              default '{}'::json not null,
    constraint fk_asset_merchant_list
        foreign key (merchant_owner)
            references merchant_list (id)
);
create unique index if not exists assets_issuer on assets (blockchain ASC, code ASC, issuer ASC);

create table if not exists merchant_assets
(
    id               bigserial primary key,
    created_at       timestamp with time zone not null,
    updated_at       timestamp with time zone not null,
    deleted_at       timestamp with time zone,
    asset_id         bigint                   not null,
    merchant_list_id bigint                   not null,
    meta_data        json default '{}'::json  not null,
    constraint fk_merchant_asset_assets
        foreign key (asset_id)
            references assets (id),
    constraint fk_merchant_users_merchant_list
        foreign key (merchant_list_id)
            references merchant_list (id)
);


