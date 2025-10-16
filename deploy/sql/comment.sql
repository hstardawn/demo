create table comment
(
    id         bigint               not null
        primary key,
    post_id    bigint               null comment '帖子id',
    user_id    bigint               not null comment '评论者ID',
    content    varchar(100)         not null comment '评论内容',
    parent_id  bigint               null comment '父评论ID',
    is_deleted tinyint(1) default 0 not null,
    ctime      bigint               not null comment '创建时间',
    utime      bigint               not null comment '修改时间'
);

