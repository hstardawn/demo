create table user
(
    id       bigint auto_increment comment '自增ID'
        primary key,
    username varchar(20)  not null comment '用户名',
    password varchar(255) not null comment '密码',
    avatar   varchar(500) not null comment '头像',
    name     varchar(20)  not null,
    ctime    bigint       not null comment '创建时间',
    utime    bigint       not null comment '修改时间',
    constraint uni_username
        unique (username)
);

