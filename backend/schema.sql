-- 实验室设备预约与共享系统 数据库结构定义 (MySQL 8.0 DDL)
-- 字符集: utf8mb4, 排序规则: utf8mb4_general_ci

CREATE DATABASE IF NOT EXISTS `lab_system` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
USE `lab_system`;

-- 1. 课题组表
CREATE TABLE IF NOT EXISTS `research_groups` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(100) NOT NULL UNIQUE COMMENT '课题组全称',
    `leader_id` BIGINT UNSIGNED DEFAULT 0 COMMENT '课题组长用户ID',
    `leader_name` VARCHAR(50) DEFAULT '' COMMENT '组长姓名',
    `department` VARCHAR(100) DEFAULT '' COMMENT '所属院系',
    `balance` DECIMAL(12, 2) DEFAULT 10000.00 COMMENT '虚拟科研经费余额(元)',
    `total_hours_used` DECIMAL(10, 2) DEFAULT 0.00 COMMENT '累计机时(小时)',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='课题组与虚拟账户表';

-- 2. 用户表 (RBAC)
CREATE TABLE IF NOT EXISTS `users` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `username` VARCHAR(50) NOT NULL UNIQUE COMMENT '学号/教工号',
    `password` VARCHAR(100) NOT NULL COMMENT '登录密码',
    `real_name` VARCHAR(50) NOT NULL COMMENT '真实姓名',
    `email` VARCHAR(100) DEFAULT '' COMMENT '电子邮箱',
    `phone` VARCHAR(20) DEFAULT '' COMMENT '手机号',
    `role` VARCHAR(20) NOT NULL DEFAULT 'student' COMMENT '角色: student, advisor, equipment_admin, system_admin',
    `group_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '所属课题组ID',
    `group_name` VARCHAR(100) DEFAULT '' COMMENT '课题组名称',
    `credit_score` INT DEFAULT 100 COMMENT '科研履约信誉积分(0-100)',
    `is_frozen` TINYINT(1) DEFAULT 0 COMMENT '是否因失信被冻结',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX `idx_users_role` (`role`),
    INDEX `idx_users_group` (`group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户与权限表';

-- 3. 实验设备台账表
CREATE TABLE IF NOT EXISTS `equipments` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(100) NOT NULL COMMENT '设备名称',
    `model` VARCHAR(100) DEFAULT '' COMMENT '规格型号',
    `category` VARCHAR(50) DEFAULT '' COMMENT '设备分类',
    `location` VARCHAR(100) NOT NULL COMMENT '安放物理位置',
    `lab_room` VARCHAR(50) DEFAULT '' COMMENT '房间号',
    `map_x` INT DEFAULT 0 COMMENT '2D数字孪生地图X坐标(%)',
    `map_y` INT DEFAULT 0 COMMENT '2D数字孪生地图Y坐标(%)',
    `safety_level` INT DEFAULT 1 COMMENT '安全等级: 1基础, 2精密, 3危险/特种',
    `training_required` TINYINT(1) DEFAULT 0 COMMENT '是否需操作认证',
    `hourly_rate` DECIMAL(10, 2) DEFAULT 0.00 COMMENT '小时单价(元/小时)',
    `status` VARCHAR(20) DEFAULT 'available' COMMENT '状态: available, reserved, in_use, maintenance, faulty, scrapped',
    `approval_type` VARCHAR(20) DEFAULT 'single' COMMENT '审批流策略: none免审, single管理员, dual导师+管理员',
    `group_id` BIGINT UNSIGNED DEFAULT 0 COMMENT '出资课题组ID',
    `group_name` VARCHAR(100) DEFAULT '' COMMENT '出资课题组名称',
    `admin_id` BIGINT UNSIGNED DEFAULT 0 COMMENT '负责管理员ID',
    `admin_name` VARCHAR(50) DEFAULT '' COMMENT '管理员姓名',
    `health_score` INT DEFAULT 100 COMMENT '健康度得分(0-100)',
    `total_hours` DECIMAL(10, 2) DEFAULT 0.00 COMMENT '累计机时',
    `max_temp` INT DEFAULT 0 COMMENT '最高工况温度(℃)',
    `vacuum_required` TINYINT(1) DEFAULT 0 COMMENT '是否需真空抽气',
    `support_batching` TINYINT(1) DEFAULT 1 COMMENT '是否支持伴热拼单优惠',
    `specs` TEXT COMMENT '技术指标',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX `idx_eq_status` (`status`),
    INDEX `idx_eq_category` (`category`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='设备全生命周期台账表';

-- 4. 预约工单表 (防重叠时间区间核心)
CREATE TABLE IF NOT EXISTS `reservations` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `equipment_id` BIGINT UNSIGNED NOT NULL COMMENT '设备ID',
    `equipment_name` VARCHAR(100) DEFAULT '',
    `equipment_model` VARCHAR(100) DEFAULT '',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '预约人ID',
    `user_name` VARCHAR(50) DEFAULT '',
    `user_phone` VARCHAR(20) DEFAULT '',
    `group_id` BIGINT UNSIGNED NOT NULL COMMENT '扣费课题组ID',
    `group_name` VARCHAR(100) DEFAULT '',
    `start_time` DATETIME NOT NULL COMMENT '预约开始时间',
    `end_time` DATETIME NOT NULL COMMENT '预约结束时间',
    `duration_hours` DECIMAL(8, 2) DEFAULT 0.00 COMMENT '预约时长',
    `experiment_name` VARCHAR(100) NOT NULL COMMENT '课题实验名称',
    `purpose` TEXT COMMENT '用途简述',
    `target_temp` INT DEFAULT 0 COMMENT '工况温度',
    `vacuum_needed` TINYINT(1) DEFAULT 0,
    `is_batched` TINYINT(1) DEFAULT 0 COMMENT '是否伴热拼单',
    `discount_rate` DECIMAL(4, 2) DEFAULT 1.00 COMMENT '折扣率',
    `total_cost` DECIMAL(10, 2) DEFAULT 0.00 COMMENT '费用',
    `status` VARCHAR(30) NOT NULL DEFAULT 'pending_admin' COMMENT '工单状态',
    `reject_reason` VARCHAR(255) DEFAULT '',
    `checkin_code` VARCHAR(10) DEFAULT '' COMMENT '6位动态签到核销码',
    `checkin_at` DATETIME DEFAULT NULL,
    `checkout_at` DATETIME DEFAULT NULL,
    `condition_check_in` VARCHAR(255) DEFAULT '',
    `condition_check_out` VARCHAR(255) DEFAULT '',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX `idx_res_eq_time` (`equipment_id`, `start_time`, `end_time`),
    INDEX `idx_res_user` (`user_id`),
    INDEX `idx_res_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='设备预约排期工单表';

-- 5. 审批日志表
CREATE TABLE IF NOT EXISTS `approval_logs` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `reservation_id` BIGINT UNSIGNED NOT NULL,
    `approver_id` BIGINT UNSIGNED NOT NULL,
    `approver_name` VARCHAR(50) DEFAULT '',
    `approver_role` VARCHAR(20) DEFAULT '',
    `action` VARCHAR(20) NOT NULL COMMENT 'approve 或 reject',
    `comment` VARCHAR(255) DEFAULT '',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX `idx_appr_res` (`reservation_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='多级审批日志表';

-- 6. 课题组虚拟经费流水表
CREATE TABLE IF NOT EXISTS `transaction_records` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `group_id` BIGINT UNSIGNED NOT NULL,
    `group_name` VARCHAR(100) DEFAULT '',
    `reservation_id` BIGINT UNSIGNED DEFAULT 0,
    `amount` DECIMAL(12, 2) NOT NULL COMMENT '变动金额',
    `balance_after` DECIMAL(12, 2) NOT NULL COMMENT '变动后余额',
    `type` VARCHAR(50) DEFAULT '' COMMENT '类型: reservation_fee, recharge, refund',
    `description` VARCHAR(255) DEFAULT '',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX `idx_trans_group` (`group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='课题组虚拟结算流水表';

-- 7. 科研履约信用记录表
CREATE TABLE IF NOT EXISTS `credit_records` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `user_id` BIGINT UNSIGNED NOT NULL,
    `user_name` VARCHAR(50) DEFAULT '',
    `reservation_id` BIGINT UNSIGNED DEFAULT 0,
    `delta_score` INT NOT NULL COMMENT '分值变动',
    `score_after` INT NOT NULL COMMENT '变动后分值',
    `reason` VARCHAR(255) NOT NULL,
    `type` VARCHAR(50) DEFAULT '',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX `idx_cred_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='科研履约信用奖惩流水表';

-- 8. 设备维保检修记录表
CREATE TABLE IF NOT EXISTS `maintenance_records` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `equipment_id` BIGINT UNSIGNED NOT NULL,
    `equipment_name` VARCHAR(100) DEFAULT '',
    `operator_id` BIGINT UNSIGNED DEFAULT 0,
    `operator_name` VARCHAR(50) DEFAULT '',
    `type` VARCHAR(50) NOT NULL,
    `description` TEXT,
    `cost` DECIMAL(10, 2) DEFAULT 0.00,
    `health_score_before` INT DEFAULT 100,
    `health_score_after` INT DEFAULT 100,
    `start_time` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `end_time` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX `idx_maint_eq` (`equipment_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='设备维护检修记录表';
