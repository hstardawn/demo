create table post
(
    id         bigint               not null comment '帖子ID'
        primary key,
    user_id    bigint               not null comment '用户ID',
    content    varchar(100)         not null comment '内容',
    is_deleted tinyint(1) default 0 not null,
    image_urls varchar(500)         not null comment '图片路径',
    ctime      bigint               not null comment '创建时间',
    utime      bigint               not null comment '修改时间'
);

