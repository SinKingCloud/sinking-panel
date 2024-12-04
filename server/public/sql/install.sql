create table cloud_configs
(
    key         varchar(50) not null
        constraint pk_key
            unique,
    value       TEXT,
    update_time TEXT,
    create_time TEXT        not null
);
