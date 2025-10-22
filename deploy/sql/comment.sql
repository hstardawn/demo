create table comment
(
    id         bigint auto_increment
        primary key,
    post_id    bigint                                    null comment '帖子id',
    user_id    bigint                                    not null comment '评论者ID',
    content    varchar(100)                              not null comment '评论内容',
    parent_id  bigint                                    null comment '父评论ID',
    deleted_at bigint       default 0                    not null comment '删除时间(软删除)',
    created_at timestamp(3) default CURRENT_TIMESTAMP(3) not null on update CURRENT_TIMESTAMP(3) comment '创建时间',
    utime      bigint                                    not null comment '修改时间'
);

