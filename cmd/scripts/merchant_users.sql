create table if not exists merchant_users
(
    id               bigserial primary key,
    created_at       timestamp with time zone not null,
    updated_at       timestamp with time zone not null,
    deleted_at       timestamp with time zone,
    user_id          bigint                   not null,
    merchant_list_id bigint                   not null,
    access           json default '{}'::json  not null,
    meta_data        json default '{}'::json  not null,
    constraint fk_merchant_users_user
        foreign key (user_id)
            references users (id),
    constraint fk_merchant_users_merchant_list
        foreign key (merchant_list_id)
            references merchant_list (id)
);


