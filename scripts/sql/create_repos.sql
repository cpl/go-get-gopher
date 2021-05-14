create table repos
(
    repo_name text not null
        constraint repos_pk
            primary key,
    repo_url text not null,
    page integer not null,
    page_index integer not null
);

create index repos_page_index
    on repos (page);

create unique index repos_repo_name_uindex
    on repos (repo_name);

