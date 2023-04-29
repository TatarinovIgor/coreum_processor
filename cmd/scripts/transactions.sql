create table transactions
(
    id          bigserial
        primary key,
    guid        uuid                                           not null
        constraint trx_id_uq
            unique,
    created_at  timestamp with time zone                       not null,
    updated_at  timestamp with time zone                       not null,
    deleted_at  timestamp with time zone,
    merchant_id varchar(64)                                    not null,
    external_id varchar(64)                                    not null,
    blockchain  varchar(32)                                    not null,
    action      varchar(32)                                    not null,
    ext_wallet  varchar                                        not null,
    status      varchar(32)                                    not null,
    asset       varchar(32)                                    not null,
    issuer      varchar                                        not null,
    amount      double precision default 0.0                   not null,
    commission  double precision default 0.0                   not null,
    hash1       varchar          default ''::character varying not null,
    hash2       varchar          default ''::character varying not null,
    hash3       varchar          default ''::character varying not null,
    hash4       varchar          default ''::character varying not null,
    hash5       varchar          default ''::character varying not null,
    callback    varchar          default ''::character varying not null
);

alter table transactions
    owner to postgres;

create index trx_guid_idx
    on transactions (guid);

