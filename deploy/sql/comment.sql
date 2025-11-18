CREATE TABLE `comment` (
                           `id` bigint NOT NULL AUTO_INCREMENT,
                           `post_id` bigint DEFAULT NULL COMMENT '帖子id',
                           `user_id` bigint NOT NULL COMMENT '评论者ID',
                           `content` varchar(100) NOT NULL COMMENT '评论内容',
                           `parent_id` bigint DEFAULT NULL COMMENT '父评论ID',
                           `deleted_at` bigint NOT NULL DEFAULT '0' COMMENT '删除时间(软删除)',
                           `created_at` timestamp(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '创建时间',
                           `utime` bigint NOT NULL COMMENT '修改时间',
                           PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci

