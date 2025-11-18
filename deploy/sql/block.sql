CREATE TABLE `block` (
                         `id` bigint NOT NULL AUTO_INCREMENT,
                         `user_id` bigint NOT NULL COMMENT '用户ID',
                         `blocked_id` bigint NOT NULL COMMENT '被拉黑的用户ID',
                         `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '拉黑状态',
                         `created_at` timestamp(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
                         `updated_at` timestamp(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '修改时间',
                         PRIMARY KEY (`id`),
                         UNIQUE KEY `block_pk` (`user_id`,`blocked_id`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci

