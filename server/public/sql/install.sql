create table cloud_configs
(
    key         varchar(50) not null
        constraint pk_key
            unique,
    value       TEXT,
    update_time TEXT,
    create_time TEXT        not null
);

create table cloud_logs
(
    id          bigint            not null
        constraint pk_id
            primary key,
    type        integer default 0 not null,
    ip          varchar(50),
    title       varchar(200),
    content     text,
    update_time text,
    create_time text
);

create index cloud_logs_createTime_index
    on cloud_logs (create_time);

create index cloud_logs_ip_index
    on cloud_logs (ip);

create index cloud_logs_type_index
    on cloud_logs (type);

create index cloud_logs_updateTime_index
    on cloud_logs (update_time);

create table cloud_servers
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

create index cloud_servers_authType_index
    on cloud_servers (auth_type);

create index cloud_servers_createTime_index
    on cloud_servers (create_time);

create index cloud_servers_ip_index
    on cloud_servers (ip);

create index cloud_servers_port_index
    on cloud_servers (port);

create index cloud_servers_updateTime_index
    on cloud_servers (update_time);

create index cloud_servers_user_index
    on cloud_servers (user);

create table cloud_tasks
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

create index cloud_tasks_createTime_index
    on cloud_tasks (create_time);

create index cloud_tasks_status_index
    on cloud_tasks (status);

create index cloud_tasks_type_index
    on cloud_tasks (type);

create index cloud_tasks_updateTime_index
    on cloud_tasks (update_time);






