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
    type        integer default 0 not null,
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

create index if not exists cloud_tasks_type_index
    on cloud_tasks (type);

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
    sort        integer default 0 not null,
    update_time text,
    create_time text
);

create index if not exists cloud_scripts_createTime_index
    on cloud_scripts (create_time);

create index if not exists cloud_scripts_typeId_sort_index
    on cloud_scripts (type_id, sort);

create index if not exists cloud_scripts_updateTime_index
    on cloud_scripts (update_time);
