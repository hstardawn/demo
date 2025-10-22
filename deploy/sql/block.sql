create table block
(
    id         bigint auto_increment
        primary key,
    user_id    bigint                                    not null comment '用户ID',
    blocked_id bigint                                    not null comment '被拉黑的用户ID',
    created_at timestamp(3) default CURRENT_TIMESTAMP(3) not null comment '创建时间',
    updated_at timestamp(3) default CURRENT_TIMESTAMP(3) not null on update CURRENT_TIMESTAMP(3) comment '修改时间',
    status     tinyint(1)   default 1                    not null comment '拉黑状态'
);

