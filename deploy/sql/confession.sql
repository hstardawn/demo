CREATE TABLE `confession` (
                              `id` bigint NOT NULL AUTO_INCREMENT COMMENT '帖子ID',
                              `user_id` bigint NOT NULL COMMENT '用户ID',
                              `name` varchar(15) NOT NULL COMMENT '昵称',
                              `content` varchar(100) NOT NULL COMMENT '内容',
                              `image_urls` varchar(500) NOT NULL COMMENT '图片路径',
                              `is_anonymous` tinyint(1) NOT NULL DEFAULT '0' COMMENT '匿名',
                              `is_visible` tinyint(1) NOT NULL DEFAULT '1' COMMENT '可见性',
                              `created_at` timestamp(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '创建时间',
                              `updated_at` timestamp(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '修改时间',
                              `deleted_at` bigint NOT NULL DEFAULT '0' COMMENT '删除时间(软删除)',
                              PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=26 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci

