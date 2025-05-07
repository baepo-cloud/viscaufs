-- migrate:up transaction:false
PRAGMA journal_mode = WAL;

alter table image_layers
    add column position integer default 0;



-- migrate:down transaction:false

PRAGMA journal_mode = WAL;

alter table image_layers
    drop column position;

