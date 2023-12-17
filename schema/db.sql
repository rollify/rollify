CREATE TABLE IF NOT EXISTS `room`
(
    `id` VARCHAR(255) NOT NULL,
    `created_at` DATETIME NOT NULL,
    `name` VARCHAR(255) NOT NULL,
    
    PRIMARY KEY(`id`)

) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;

CREATE TABLE IF NOT EXISTS  `user`
(
    `id` VARCHAR(255) NOT NULL,
    `created_at` DATETIME(3) NOT NULL,
    `name` VARCHAR(255) NOT NULL COLLATE utf8mb4_0900_ai_ci,
    `room_id` VARCHAR(255) NOT NULL,

    PRIMARY KEY(`id`),

    INDEX `idx_user_name_room_id` (`name`, `room_id`)

) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;

CREATE TABLE IF NOT EXISTS  `dice_roll`
(
    `id` VARCHAR(255) NOT NULL,
    `created_at` DATETIME(3) NOT NULL,
    `serial` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT UNIQUE,
    `user_id` VARCHAR(255) NOT NULL,
    `room_id` VARCHAR(255) NOT NULL,

    PRIMARY KEY(`id`),

    INDEX `idx_room_id` (`room_id`)

) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;

CREATE TABLE IF NOT EXISTS  `die_roll`
(
    `id` VARCHAR(255) NOT NULL,
    `dice_roll_id` VARCHAR(255) NOT NULL,
    `die_type_id` VARCHAR(255) NOT NULL,
    `side` SMALLINT UNSIGNED NOT NULL,

    PRIMARY KEY(`id`),

    INDEX `idx_dice_roll_id` (`dice_roll_id`)

) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
