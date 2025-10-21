create table block
(
    id         bigint auto_increment
        primary key,
    user_id    bigint               not null comment '用户ID',
    blocked_id bigint               not null comment '被拉黑的用户ID',
    ctime      bigint               not null comment '创建时间',
    utime      bigint               not null comment '修改时间',
    status     tinyint(1) default 1 not null comment '拉黑状态'
);

