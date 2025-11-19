CREATE TABLE `comment` (
                           `id` bigint NOT NULL AUTO_INCREMENT,
                           `confession_id` bigint DEFAULT NULL COMMENT '表白帖子id',
                           `user_id` bigint NOT NULL COMMENT '评论者ID',
                           `content` varchar(100) NOT NULL COMMENT '评论内容',
                           `parent_id` bigint DEFAULT NULL COMMENT '父评论ID',
                           `created_at` timestamp(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '创建时间',
                           `updated_at` timestamp(3) NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
                           `deleted_at` bigint NOT NULL DEFAULT '0' COMMENT '删除时间(软删除)',
                           PRIMARY KEY (`id`),
                           KEY `comment_confession_id_index` (`confession_id`),
                           KEY `comment_parent_id_index` (`parent_id`) COMMENT 'parent_id'
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci

