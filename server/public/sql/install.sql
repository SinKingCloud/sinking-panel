create table if not exists cloud_configs
(
    key         varchar(50) not null
        constraint pk_key
            unique,
    value       TEXT,
    update_time TEXT,
    create_time TEXT        not null
);

create table if not exists cloud_logs
(
    id          bigint            not null
        constraint pk_id
            primary key,
    type        integer default 0 not null,
    ip          varchar(50),
    location    varchar(100),
    title       varchar(200),
    content     text,
    update_time text,
    create_time text
);

create index if not exists cloud_logs_createTime_index
    on cloud_logs (create_time);

create index if not exists cloud_logs_ip_index
    on cloud_logs (ip);

create index if not exists cloud_logs_type_index
    on cloud_logs (type);

create index if not exists cloud_logs_updateTime_index
    on cloud_logs (update_time);

create table if not exists cloud_servers
(
    id          bigint            not null
        constraint pk_id
            primary key,
    ip          varchar(100)      not null,
    port        integer           not null,
    user        varchar(50),
    auth_type   integer default 0 not null,
    password    text,
    name        varchar(100),
    update_time text,
    create_time text
);

create index if not exists cloud_servers_authType_index
    on cloud_servers (auth_type);

create index if not exists cloud_servers_createTime_index
    on cloud_servers (create_time);

create index if not exists cloud_servers_ip_index
    on cloud_servers (ip);

create index if not exists cloud_servers_port_index
    on cloud_servers (port);

create index if not exists cloud_servers_updateTime_index
    on cloud_servers (update_time);

create index if not exists cloud_servers_user_index
    on cloud_servers (user);

create table if not exists cloud_tasks
(
    id          bigint            not null
        constraint cloud_tasks_pk_id
            primary key,
    entry_id    integer default 0 not null,
    name        varchar(50)       not null,
    spec        varchar(50)       not null,
    type_id     bigint  default 0 not null,
    exec_type   integer default 0 not null,
    script      text              not null,
    run_time    text,
    status      integer default 0 not null,
    update_time text,
    create_time text
);

create index if not exists cloud_tasks_createTime_index
    on cloud_tasks (create_time);

create index if not exists cloud_tasks_status_index
    on cloud_tasks (status);

create index if not exists cloud_tasks_execType_index
    on cloud_tasks (exec_type);

create index if not exists cloud_tasks_typeId_index
    on cloud_tasks (type_id);

create index if not exists cloud_tasks_updateTime_index
    on cloud_tasks (update_time);

create table if not exists cloud_types
(
    id          bigint            not null
        constraint cloud_types_pk_id
            primary key,
    module      varchar(50)       not null,
    name        varchar(50)       not null,
    sort        integer default 0 not null,
    update_time text,
    create_time text
);

create index if not exists cloud_types_createTime_index
    on cloud_types (create_time);

create index if not exists cloud_types_module_name_index
    on cloud_types (module, name);

create index if not exists cloud_types_module_sort_index
    on cloud_types (module, sort);

create index if not exists cloud_types_updateTime_index
    on cloud_types (update_time);

create table if not exists cloud_scripts
(
    id          bigint            not null
        constraint cloud_scripts_pk_id
            primary key,
    type_id     bigint  default 0 not null,
    name        varchar(100)      not null,
    script      text              not null,
    update_time text,
    create_time text
);

create index if not exists cloud_scripts_createTime_index
    on cloud_scripts (create_time);

create index if not exists cloud_scripts_typeId_index
    on cloud_scripts (type_id);

create index if not exists cloud_scripts_updateTime_index
    on cloud_scripts (update_time);

create table if not exists cloud_sites
(
    id          bigint            not null
        constraint cloud_sites_pk_id
            primary key,
    type_id     bigint  default 0 not null,
    name        varchar(100)      not null,
    type        integer default 0 not null,
    status      integer default 0 not null,
    root        text    default '' not null,
    run_path    text    default '' not null,
    config      text    default '' not null,
    update_time text,
    create_time text
);

create index if not exists cloud_sites_createTime_index
    on cloud_sites (create_time);

create index if not exists cloud_sites_status_index
    on cloud_sites (status);

create index if not exists cloud_sites_typeId_index
    on cloud_sites (type_id);

create index if not exists cloud_sites_type_index
    on cloud_sites (type);

create index if not exists cloud_sites_updateTime_index
    on cloud_sites (update_time);

create table if not exists cloud_secrets
(
    id          bigint       not null
        constraint cloud_secrets_pk_id
            primary key,
    name        varchar(100) not null,
    provider    integer default 0 not null,
    data        text         not null,
    update_time text,
    create_time text
);

create index if not exists cloud_secrets_name_index
    on cloud_secrets (name);

create index if not exists cloud_secrets_provider_index
    on cloud_secrets (provider);

create index if not exists cloud_secrets_createTime_index
    on cloud_secrets (create_time);

create index if not exists cloud_secrets_updateTime_index
    on cloud_secrets (update_time);

create table if not exists cloud_certs
(
    id            bigint                 not null
        constraint cloud_certs_pk_id
            primary key,
    name          varchar(100)           not null,
    type          integer default 0      not null,
    challenge     varchar(20) default '' not null,
    secret_id     bigint default 0       not null,
    auto_renew    integer default 0      not null,
    domains       text    default ''     not null,
    certificate   text                   not null,
    private_key   text                   not null,
    start_time    text,
    expire_time   text,
    update_time   text,
    create_time   text
);

create index if not exists cloud_certs_createTime_index
    on cloud_certs (create_time);

create index if not exists cloud_certs_expireTime_index
    on cloud_certs (expire_time);

create index if not exists cloud_certs_type_index
    on cloud_certs (type);

create index if not exists cloud_certs_secretId_index
    on cloud_certs (secret_id);

create index if not exists cloud_certs_autoRenew_expireTime_index
    on cloud_certs (auto_renew, expire_time);

create index if not exists cloud_certs_updateTime_index
    on cloud_certs (update_time);

create table if not exists cloud_site_domains
(
    id          bigint            not null
        constraint cloud_site_domains_pk_id
            primary key,
    site_id     bigint            not null,
    cert_id     bigint  default 0 not null,
    domain      varchar(255)      not null collate nocase,
    update_time text,
    create_time text
);

create index if not exists cloud_site_domains_createTime_index
    on cloud_site_domains (create_time);

create index if not exists cloud_site_domains_domain_index
    on cloud_site_domains (domain);

create index if not exists cloud_site_domains_siteId_index
    on cloud_site_domains (site_id);

create index if not exists cloud_site_domains_certId_index
    on cloud_site_domains (cert_id);

create index if not exists cloud_site_domains_updateTime_index
    on cloud_site_domains (update_time);
