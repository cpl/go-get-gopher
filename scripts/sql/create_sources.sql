create table sources
(
    repo_name text not null
        references repos,
    url text not null,
    domain text,
    ext text
);

create index sources_domain_index
    on sources (domain);

create index sources_ext_index
    on sources (ext);

create index sources_repo_name_index
    on sources (repo_name);
