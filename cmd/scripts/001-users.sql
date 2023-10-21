create table if not exists users
(
    id                   bigserial
        constraint users_pk
            primary key,
    created_at           timestamp with time zone not null,
    updated_at           timestamp with time zone not null,
    deleted_at           timestamp with time zone,
    identity             varchar(255)             not null
        constraint users_pk_guid
            unique,
    first_name           varchar(255)             not null,
    last_name            varchar(255)             not null,
    terms_and_conditions boolean,
    access               integer,
    meta_data            json default '{}'::json
);