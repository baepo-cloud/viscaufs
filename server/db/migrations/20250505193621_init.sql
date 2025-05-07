-- migrate:up transaction:false

PRAGMA journal_mode = WAL;

create table images
(
    id           text primary key,
    repository   text      not null,
    identifier   text      not null,
    digest       text      not null,
    layers_count integer   not null,
    manifest     text      not null,
    created_at   timestamp not null default CURRENT_TIMESTAMP
);

create table layers
(
    id         text primary key,
    digest     text      not null,
    created_at timestamp not null default CURRENT_TIMESTAMP
);

create table image_layers
(
    image_id   text      not null,
    layer_id   text      not null,
    created_at timestamp not null default CURRENT_TIMESTAMP,
    primary key (image_id, layer_id),
    foreign key (image_id) references images (id),
    foreign key (layer_id) references layers (id)
);

create unique index images_digest_idx on images (digest);
create unique index layers_digest_idx on layers (digest);

-- migrate:down transaction:false
PRAGMA journal_mode = WAL;

drop table images;
drop table layers;
