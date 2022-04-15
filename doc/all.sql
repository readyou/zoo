DROP TABLE zoo_tab;
CREATE TABLE zoo_tab
(
    id              bigint auto_increment,
    username        varchar(64)     not null comment '用户名（全局唯一）',
    hashed_password varchar(64)     not null comment '加密之后的密码',
    nickname        varchar(64)     not null default '' comment '昵称',
    avatar          varchar(512)    not null default '' comment '头像URL',
    create_time     BIGINT UNSIGNED not null comment '创建时间',
    update_time     BIGINT UNSIGNED not null comment '更新时间',
    primary key (id),
    unique index uidx_username (username)
)