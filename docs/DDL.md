# User Service (用户服务) DDL
```sql
-- 用户基本信息表
CREATE TABLE `user` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `email` VARCHAR(100) NOT NULL COMMENT '登录邮箱，唯一',
    `password_hash` VARCHAR(255) NOT NULL COMMENT '密码哈希值',
    `nickname` VARCHAR(50) NOT NULL DEFAULT '新用户' COMMENT '用户昵称',
    `avatar_url` VARCHAR(255) COMMENT '头像OSS链接',
    `is_premium` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '是否为付费用户 (0: 否, 1: 是)',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户基本信息表';

-- 用户点数表
CREATE TABLE `user_point` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `user_id` BIGINT NOT NULL COMMENT '用户ID (逻辑外键: user.id)',
    `current_points` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '当前可用点数',
    `total_consumed` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '历史总消耗点数',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户点数表';

-- 点数交易流水表
CREATE TABLE `point_transaction` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `user_id` BIGINT NOT NULL COMMENT '用户ID (逻辑外键: user.id)',
    `type` ENUM('CONSUME', 'RECHARGE') NOT NULL COMMENT '交易类型: CONSUME-消耗, RECHARGE-充值',
    `amount` INT UNSIGNED NOT NULL COMMENT '点数变动数量',
    `related_book_id` BIGINT COMMENT '关联的绘本ID (逻辑外键: book.id), 仅消耗时可能关联',
    `description` VARCHAR(255) COMMENT '交易描述',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='点数交易流水表';
```

# Book Service (绘本服务) DDL
```sql
-- 绘本元数据表 (已添加 summary 字段)
CREATE TABLE `book` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `user_id` BIGINT NOT NULL COMMENT '作者用户ID (逻辑外键: user.id)',
    `title` VARCHAR(128) NOT NULL COMMENT '绘本标题',
    `summary` TEXT COMMENT '绘本故事内容大纲/摘要',
    `cover_url` VARCHAR(255) COMMENT '绘本封面OSS链接',
    `initial_prompt` TEXT COMMENT '初始生成绘本的Prompt',
    `style` VARCHAR(50) COMMENT '绘本风格',
    `status` ENUM('DRAFT', 'PENDING_REVIEW', 'SHARED', 'REJECTED') NOT NULL DEFAULT 'DRAFT' COMMENT '绘本状态',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `shared_at` DATETIME COMMENT '分享到社区的时间',
    PRIMARY KEY (`id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_status_shared` (`status`, `shared_at`) 
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='绘本元数据表';

-- 绘本页面内容表
CREATE TABLE `book_page` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `book_id` BIGINT NOT NULL COMMENT '所属绘本ID (逻辑外键: book.id)',
    `page_number` INT UNSIGNED NOT NULL COMMENT '页码，从1开始',
    `text_content` TEXT NOT NULL COMMENT '页面文本内容',
    `image_url` VARCHAR(255) COMMENT '页面图片OSS链接',
    `image_prompt` TEXT COMMENT '用于图片生成的Prompt',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_book_page` (`book_id`, `page_number`),
    KEY `idx_book_id` (`book_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='绘本页面内容表';

-- AI 任务表
CREATE TABLE `ai_generation_task` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `user_id` BIGINT NOT NULL COMMENT '提交用户ID (逻辑外键: user.id)',
    `book_id` BIGINT COMMENT '关联的绘本ID (逻辑外键: book.id)',
    `page_id` BIGINT COMMENT '关联的页面ID (逻辑外键: book_page.id)',
    `task_type` ENUM('FULL_BOOK', 'PAGE_REGEN') NOT NULL COMMENT '任务类型: FULL_BOOK-完整绘本生成, PAGE_REGEN-单页图片重绘',
    `status` ENUM('PENDING', 'PROCESSING', 'COMPLETED', 'FAILED') NOT NULL DEFAULT 'PENDING' COMMENT '任务状态',
    `submit_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '任务提交时间',
    `finish_time` DATETIME COMMENT '任务完成时间',
    `consumed_points` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '消耗的点数',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_status_submit` (`status`, `submit_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI 生成任务表';
```

# Community & Interaction (社区互动表) DDL
```sql
-- 社区绘本统计表
CREATE TABLE `community_book_stat` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `book_id` BIGINT NOT NULL COMMENT '绘本ID (逻辑外键: book.id)',
    `like_count` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '点赞数',
    `collect_count` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '收藏数',
    `comment_count` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '评论数',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_book_id` (`book_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='社区绘本统计表';

-- 绘本点赞表
CREATE TABLE `book_like` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `user_id` BIGINT NOT NULL COMMENT '用户ID (逻辑外键: user.id)',
    `book_id` BIGINT NOT NULL COMMENT '绘本ID (逻辑外键: book.id)',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '点赞时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_book` (`user_id`, `book_id`),
    KEY `idx_book_id` (`book_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='绘本点赞表';

-- 绘本收藏表
CREATE TABLE `book_collection` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `user_id` BIGINT NOT NULL COMMENT '用户ID (逻辑外键: user.id)',
    `book_id` BIGINT NOT NULL COMMENT '绘本ID (逻辑外键: book.id)',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '收藏时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_book` (`user_id`, `book_id`),
    KEY `idx_book_id` (`book_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='绘本收藏表';

-- 绘本评论表
CREATE TABLE `book_comment` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `book_id` BIGINT NOT NULL COMMENT '所属绘本ID (逻辑外键: book.id)',
    `user_id` BIGINT NOT NULL COMMENT '评论用户ID (逻辑外键: user.id)',
    `parent_id` BIGINT COMMENT '父评论ID，用于回复 (逻辑外键: book_comment.id)',
    `content` TEXT NOT NULL COMMENT '评论内容',
    `status` ENUM('APPROVED', 'PENDING', 'REJECTED') NOT NULL DEFAULT 'PENDING' COMMENT '评论审核状态',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_book_id` (`book_id`),
    KEY `idx_parent_id` (`parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='绘本评论表';
```

# Audit & Security (审核与安全表) DDL
```sql
-- 内容审核日志表
CREATE TABLE `content_audit_log` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `content_type` ENUM('BOOK', 'PAGE_TEXT', 'PAGE_IMAGE', 'COMMENT') NOT NULL COMMENT '内容类型',
    `content_id` BIGINT NOT NULL COMMENT '被审核内容的ID (逻辑外键)',
    `submit_user_id` BIGINT NOT NULL COMMENT '内容提交用户ID (逻辑外键: user.id)',
    `audit_result` ENUM('PASS', 'REJECT', 'NEED_MANUAL') NOT NULL COMMENT '审核结果',
    `trigger_type` ENUM('AUTO', 'MANUAL') NOT NULL COMMENT '触发类型',
    `details` JSON COMMENT '审核详情 (如敏感词列表, 机器分数)',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_content` (`content_type`, `content_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='内容审核日志表';

-- 举报表
CREATE TABLE `report` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `reported_user_id` BIGINT COMMENT '被举报人ID (逻辑外键: user.id)',
    `reported_content_type` ENUM('BOOK', 'COMMENT') NOT NULL COMMENT '被举报内容类型',
    `reported_content_id` BIGINT NOT NULL COMMENT '被举报内容的ID (逻辑外键)',
    `report_user_id` BIGINT NOT NULL COMMENT '举报人ID (逻辑外键: user.id)',
    `reason` TEXT NOT NULL COMMENT '举报理由',
    `status` ENUM('PENDING', 'PROCESSED') NOT NULL DEFAULT 'PENDING' COMMENT '处理状态',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_reported_content` (`reported_content_type`, `reported_content_id`),
    KEY `idx_report_user_id` (`report_user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='举报表';
```

