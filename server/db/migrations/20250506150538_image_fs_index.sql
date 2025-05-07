-- migrate:up transaction:false

PRAGMA journal_mode = WAL;

alter table images
    add column fs_index BLOB default null;

-- migrate:down transaction:false

PRAGMA journal_mode = WAL;

alter table images
    drop column fs_index;
