create table if not exists merchant_list
(
    id           bigserial
        primary key,
    created_at   timestamp with time zone                         not null,
    updated_at   timestamp with time zone                         not null,
    deleted_at   timestamp with time zone,
    email        varchar(256) default ''::character varying       not null,
    company_name varchar      default ''::character varying       not null,
    type         varchar(255) default 'crypto'::character varying not null,
    is_blocked   bool         default false                       not null,
    merchant_id  varchar(64)  default null,
    meta_data    json         default '{}'::json                  not null
);
CREATE UNIQUE INDEX if not exists merchant_id
    on merchant_list (merchant_id);
