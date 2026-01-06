create table if not exists pages (
    id serial primary key,
    url text unique not null,
    raw_html text not null,
    is_sent boolean default false,
    created_at timestamp default current_timestamp
);

-- index for kafka
create index if not exists idx_pages_not_sent on pages(id) where is_sent = false;
