-- migrate:up transaction:false
PRAGMA journal_mode = WAL;

alter table images add layer_digests text default '' not null;

create index images_digest_idx on images (layer_digests);
create index layers_digest_idx on images (layer_digests);

-- migrate:down

PRAGMA journal_mode = WAL;

alter table images drop column layer_digests;
drop index images_digest_idx;
drop index layers_digest_idx;
