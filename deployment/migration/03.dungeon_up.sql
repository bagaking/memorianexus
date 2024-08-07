CREATE TABLE `dungeons` (
    `id` BIGINT UNSIGNED NOT NULL,
    `user_id` BIGINT UNSIGNED NOT NULL,

    `type` TINYINT UNSIGNED NOT NULL COMMENT 'Dungeon (<256) 类型分段 campaign = 001, endless = 002, instance > 003',

    `title` VARCHAR(255) NOT NULL,
    `description` TEXT,
#     `rule` TEXT COMMENT '复习规则的详细配置, JSON格式',

    -- practice preference (set when create, fork default values from user's setting)
    `review_interval` VARCHAR(255) COMMENT "Interval for review in days",
    `difficulty_preference` TINYINT UNSIGNED COMMENT "User's preference for difficulty",
    `quiz_mode` VARCHAR(32) COMMENT "Preferred quiz mode",
    `priority_mode` VARCHAR(255) COMMENT "Preferred priority mode",

    -- system
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` DATETIME DEFAULT NULL COMMENT "Record delete time in UTC",

    PRIMARY KEY (`id`),
    INDEX `idx_user_dungeon` (`user_id`, `id`),
    INDEX `idx_type_user` (`type`, `user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `dungeon_books` (
    `dungeon_id` BIGINT UNSIGNED NOT NULL,
    `book_id` BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (`dungeon_id`, `book_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `dungeon_tags` (
    `dungeon_id` BIGINT UNSIGNED NOT NULL,
    `tag_id` BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (`dungeon_id`, `tag_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- monster 代表的是 item 对于特定 user 的属性，ID 就用 itemID
CREATE TABLE `monsters` (
    `user_id` BIGINT UNSIGNED NOT NULL,
    `item_id` BIGINT UNSIGNED NOT NULL,

    `familiarity` TINYINT UNSIGNED COMMENT "percentage: 0-100",

    PRIMARY KEY (`user_id`, `item_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- dungeon 和 user 是 n 对 1 的，因此 `dungeon_id` + `item_id` 是可以对应到 monster
CREATE TABLE `dungeon_monsters` (
    `dungeon_id` BIGINT UNSIGNED NOT NULL,
    `item_id` BIGINT UNSIGNED NOT NULL,

    `source_type` TINYINT UNSIGNED COMMENT "source type of the monster, item=1, book=2, tag=3",
    `source_id` BIGINT UNSIGNED NOT NULL,

    -- 宽表用途，用于查询
    `familiarity`  TINYINT UNSIGNED COMMENT "percentage: 0-100",
    `importance` TINYINT UNSIGNED COMMENT "Importance",
    `difficulty` TINYINT UNSIGNED COMMENT "Difficulty",

    -- gaming
    `visibility` TINYINT UNSIGNED NOT NULL COMMENT "percentage: 0-100",
    `avatar` VARCHAR(2048) NOT NULL COMMENT "avatar of the monster",
    `name` VARCHAR(2048) NOT NULL COMMENT "name of the monster", -- can be generated by AI
    `description` VARCHAR(2048) NOT NULL COMMENT "description of the monster", -- can be generated by AI

    -- runtime
    `practice_at` DATETIME DEFAULT NULL,
    `next_practice_at` DATETIME DEFAULT CURRENT_TIMESTAMP, -- set to current_time for initial
    `practice_count` INT UNSIGNED DEFAULT 0,

    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,



    PRIMARY KEY (`dungeon_id`, `item_id`),
    INDEX idx_practice_order (`dungeon_id`, `next_practice_at`, `familiarity`, `importance`, `difficulty`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `user_monsters` (
    `user_id` BIGINT UNSIGNED NOT NULL,
    `item_id` BIGINT UNSIGNED NOT NULL,
    `familiarity`  TINYINT UNSIGNED COMMENT "percentage: 0-100",

    PRIMARY KEY (`user_id`, `item_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
