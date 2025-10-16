create table post
(
    id           bigint auto_increment comment '帖子ID'
        primary key,
    user_id      bigint               not null comment '用户ID',
    name         varchar(15)          not null comment '昵称',
    content      varchar(100)         not null comment '内容',
    is_deleted   tinyint(1) default 0 not null,
    image_urls   varchar(500)         not null comment '图片路径',
    ctime        bigint               not null comment '创建时间',
    utime        bigint               not null comment '修改时间',
    is_anonymous tinyint(1) default 0 not null comment '匿名',
    is_visible   tinyint(1) default 1 not null comment '可见性'
);

