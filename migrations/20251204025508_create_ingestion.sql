-- +goose up
create table ingested_documents (
    id              uuid primary key,
    document_type   text not null,
    external_id     text not null,
    created_at      timestamptz not null default now(),
    updated_at      timestamptz not null default now()
);

create index idx_ingested_documents_type on ingested_documents(document_type);
create unique index idx_ingested_documents_reference on ingested_documents(document_type, external_id);

-- +goose down
drop table ingested_documents;
