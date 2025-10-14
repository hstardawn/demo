create table comment
(
    id         int                  not null
        primary key,
    post_id    int                  null comment '帖子id',
    user_id    int                  not null comment '评论者ID',
    content    varchar(100)         not null comment '评论内容',
    parent_id  int                  null comment '父评论ID',
    deleted_at tinyint(1) default 0 not null,
    visible    tinyint(1) default 1 not null comment '可见性',
    anonymous  tinyint(1) default 0 not null comment '匿名'
);

