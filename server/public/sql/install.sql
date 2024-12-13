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
    id          integer           not null
        constraint pk_id
            primary key autoincrement,
    type        integer default 0 not null,
    ip          varchar(50),
    title       varchar(200),
    content     text,
    update_time text,
    create_time text
);

create index idx_createTime
    on cloud_logs (create_time);

create index idx_ip
    on cloud_logs (ip);

create index idx_type
    on cloud_logs (type);

create index idx_updateTime
    on cloud_logs (update_time);
