-- migrate:up transaction:false

PRAGMA journal_mode = WAL;

alter table layers
    add column fs_index BLOB default null;

-- migrate:down transaction:false

PRAGMA journal_mode = WAL;

alter table layers
    drop column fs_index;

