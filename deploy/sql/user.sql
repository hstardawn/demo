create table user
(
    id         bigint auto_increment comment '自增ID'
        primary key,
    username   varchar(20)                               not null comment '用户名',
    password   varchar(255)                              not null comment '密码',
    avatar     varchar(500)                              not null comment '头像',
    name       varchar(20)                               not null,
    created_at timestamp(3) default CURRENT_TIMESTAMP(3) not null comment '创建时间',
    updated_at timestamp(3) default CURRENT_TIMESTAMP(3) not null on update CURRENT_TIMESTAMP(3) comment '修改时间',
    constraint uni_username
        unique (username)
);

