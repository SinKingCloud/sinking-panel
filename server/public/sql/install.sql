CREATE TABLE `cloud_configs` (
                                 `id` bigint(32) NOT NULL AUTO_INCREMENT COMMENT '配置ID',
                                 `ident_id` varchar(20) NOT NULL COMMENT '标识ID',
                                 `key` varchar(100) NOT NULL COMMENT '配置标识',
                                 `value` varchar(255) DEFAULT NULL COMMENT '配置内容',
                                 `create_time` datetime DEFAULT NULL COMMENT '创建时间',
                                 `update_time` datetime DEFAULT NULL COMMENT '更新时间',
                                 PRIMARY KEY (`id`),
                                 KEY `idx_identId` (`ident_id`),
                                 KEY `idx_identId_key` (`ident_id`,`key`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;

CREATE TABLE `cloud_notices` (
                                 `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '公告ID',
                                 `place` varchar(50) NOT NULL COMMENT '公告位置',
                                 `title` varchar(30) NOT NULL COMMENT '公告名称',
                                 `content` text COMMENT '公告内容',
                                 `sort` int(11) NOT NULL DEFAULT '0' COMMENT '排序',
                                 `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '公告状态(0:正常/1:禁用)',
                                 `look_num` int(11) NOT NULL DEFAULT '0' COMMENT '查看次数',
                                 `update_time` datetime DEFAULT NULL COMMENT '更新时间',
                                 `create_time` datetime DEFAULT NULL COMMENT '创建时间',
                                 PRIMARY KEY (`id`),
                                 KEY `idx_place_status_sort` (`place`,`status`,`sort`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;

CREATE TABLE `cloud_pay_logs` (
                                  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '记录ID',
                                  `user_id` int(11) NOT NULL DEFAULT '0' COMMENT '用户ID',
                                  `type` tinyint(1) NOT NULL DEFAULT '0' COMMENT '类型(0:增加/1:减少)',
                                  `title` varchar(255) NOT NULL COMMENT '标题',
                                  `content` text NOT NULL COMMENT '详细',
                                  `money` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '金额',
                                  `create_time` datetime DEFAULT NULL COMMENT '创建时间',
                                  `update_time` datetime DEFAULT NULL COMMENT '更新时间',
                                  PRIMARY KEY (`id`),
                                  KEY `idx_userId_createTime` (`user_id`,`create_time`),
                                  KEY `idx_userId_type_createTime` (`user_id`,`type`,`create_time`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;

CREATE TABLE `cloud_user_orders` (
                                   `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '记录ID',
                                   `user_id` int(11) NOT NULL DEFAULT '0' COMMENT '用户ID',
                                   `pay_type` tinyint(1) NOT NULL DEFAULT '0' COMMENT '支付方式(0:支付宝/1:微信支付/2:QQ支付)',
                                   `order_type` tinyint(1) NOT NULL DEFAULT '0' COMMENT '订单类型(0:余额充值)',
                                   `money` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '订单金额',
                                   `trade_no` varchar(100) NOT NULL COMMENT '系统订单号',
                                   `out_trade_no` varchar(100) DEFAULT NULL COMMENT '外部订单号',
                                   `name` varchar(50) NOT NULL COMMENT '订单名称',
                                   `param` text COMMENT '其他参数',
                                   `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '状态(0:未支付/1:已支付)',
                                   `create_time` datetime DEFAULT NULL COMMENT '创建时间',
                                   `update_time` datetime DEFAULT NULL COMMENT '更新时间',
                                   PRIMARY KEY (`id`),
                                   KEY `idx_userId_payType_orderType_status` (`user_id`,`pay_type`,`order_type`,`status`),
                                   KEY `idx_outTradeNo` (`out_trade_no`),
                                   KEY `idx_tradeNo` (`trade_no`),
                                   KEY `idx_userId_status` (`user_id`,`status`),
                                   KEY `idx_createTime_status` (`create_time`,`status`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;

CREATE TABLE `cloud_uin_cookies` (
                                     `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '记录ID',
                                     `uin` varchar(20) NOT NULL DEFAULT '0' COMMENT 'QQ号码',
                                     `domain` varchar(255) NOT NULL COMMENT 'cookie域名',
                                     `cookie` text COMMENT 'cookie值',
                                     `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT 'cookie状态(0:正常/1:失效)',
                                     `check_time` datetime DEFAULT NULL COMMENT '检测时间',
                                     `update_time` datetime DEFAULT NULL COMMENT '更新时间',
                                     `create_time` datetime DEFAULT NULL COMMENT '创建时间',
                                     PRIMARY KEY (`id`),
                                     KEY `idx_checkTime` (`check_time`),
                                     KEY `idx_status_domain_checkTime` (`status`,`domain`,`check_time`),
                                     KEY `idx_uin_status` (`uin`,`status`),
                                     KEY `idx_uin_cookies` (`uin`,`domain`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;

CREATE TABLE `cloud_uin_infos` (
                                   `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '记录ID',
                                   `uin` varchar(20) NOT NULL DEFAULT '0' COMMENT 'QQ号码',
                                   `nick_name` varchar(255) DEFAULT NULL COMMENT 'QQ昵称',
                                   `real_auth` int(11) NOT NULL DEFAULT '0' COMMENT '是否实名(0:未实名/ 1:已实名)',
                                   `level` int(11) NOT NULL DEFAULT '0' COMMENT 'QQ等级',
                                   `activate_day` varchar(11) NOT NULL DEFAULT '0' COMMENT '共计活跃天数',
                                   `today_day` varchar(11) NOT NULL DEFAULT '0' COMMENT '今日活跃天数',
                                   `speed_info` text COMMENT '加速详情',
                                   `update_time` datetime DEFAULT NULL COMMENT '更新时间',
                                   `create_time` datetime DEFAULT NULL COMMENT '创建时间',
                                   PRIMARY KEY (`id`),
                                   KEY `idx_uin` (`uin`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;

CREATE TABLE `cloud_uin_logs` (
                                  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '记录ID',
                                  `user_id` int(11) NOT NULL DEFAULT '0' COMMENT '用户ID',
                                  `type` tinyint(1) NOT NULL DEFAULT '0' COMMENT '日志类型',
                                  `uin` varchar(20) NOT NULL COMMENT 'QQ账号',
                                  `title` varchar(255) NOT NULL COMMENT '日志标题',
                                  `content` varchar(999) DEFAULT NULL COMMENT '日志详情',
                                  `update_time` datetime DEFAULT NULL COMMENT '更新时间',
                                  `create_time` datetime DEFAULT NULL COMMENT '创建时间',
                                  PRIMARY KEY (`id`),
                                  KEY `idx_uin_type` (`uin`,`type`),
                                  KEY `idx_userId_uin_type` (`user_id`,`uin`,`type`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;

CREATE TABLE `cloud_uin_servers` (
                                     `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '服务器ID',
                                     `pid` int(11) NOT NULL DEFAULT '0' COMMENT '父级ID',
                                     `key` varchar(50) DEFAULT NULL COMMENT '服务器标识',
                                     `name` varchar(255) NOT NULL COMMENT '服务器名称',
                                     `proxy` varchar(500) DEFAULT NULL COMMENT '代理地址',
                                     `uin_num` int(11) NOT NULL DEFAULT '0' COMMENT 'QQ数量',
                                     `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '状态(0:正常/1:禁用)',
                                     `update_time` datetime DEFAULT NULL COMMENT '更新时间',
                                     `create_time` datetime DEFAULT NULL COMMENT '创建时间',
                                     PRIMARY KEY (`id`),
                                     KEY `idx_key` (`key`),
                                     KEY `idx_pid` (`pid`),
                                     KEY `idx_status` (`status`),
                                     KEY `idx_uinNum` (`uin_num`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;

CREATE TABLE `cloud_uin_orders` (
                                    `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '记录ID',
                                    `user_id` int(11) NOT NULL DEFAULT '0' COMMENT '用户ID',
                                    `uin` varchar(20) NOT NULL DEFAULT '0' COMMENT 'QQ号码',
                                    `trade_no` bigint(20) NOT NULL DEFAULT '0' COMMENT '订单号',
                                    `money` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '订单金额',
                                    `refund_money` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '已退款金额',
                                    `name` varchar(255) CHARACTER SET utf8mb4 NOT NULL COMMENT '订单名称',
                                    `days` int(11) NOT NULL DEFAULT '0' COMMENT '天数',
                                    `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '订单状态(0:成功/1:已退单)',
                                    `update_time` datetime DEFAULT NULL COMMENT '更新时间',
                                    `create_time` datetime DEFAULT NULL COMMENT '创建时间',
                                    PRIMARY KEY (`id`),
                                    KEY `idx_userId_status_createTime` (`user_id`,`status`,`create_time`),
                                    KEY `idx_userId` (`user_id`),
                                    KEY `idx_uin` (`uin`),
                                    KEY `idx_status` (`status`),
                                    KEY `idx_createTime` (`create_time`),
                                    KEY `idx_tradeNo` (`trade_no`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


CREATE TABLE `cloud_uins` (
                              `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '记录ID',
                              `user_id` int(11) NOT NULL DEFAULT '0' COMMENT '用户ID',
                              `server_id` int(11) NOT NULL DEFAULT '0' COMMENT '服务ID',
                              `server_key` varchar(50) NOT NULL COMMENT '服务器标识',
                              `uin` varchar(20) NOT NULL DEFAULT '0' COMMENT 'QQ号码',
                              `password` varchar(32) NOT NULL COMMENT 'QQ密码',
                              `proto` tinyint(1) NOT NULL DEFAULT '0' COMMENT '使用协议(0:ipad/1:android/2:pad)',
                              `device` int(11) NOT NULL DEFAULT '0' COMMENT '设备码',
                              `token` text COMMENT '登陆token',
                              `uin_status` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'QQ状态(0:正常/1:异常)',
                              `expire_day` int(11) NOT NULL DEFAULT '0' COMMENT '挂机天数',
                              `run_day` int(11) NOT NULL DEFAULT '0' COMMENT '已挂机天数',
                              `setting` text COMMENT '挂机设置',
                              `timing` int(11) NOT NULL DEFAULT '0' COMMENT '定时挂机',
                              `run_time` datetime DEFAULT NULL COMMENT '运行时间',
                              `check_time` datetime DEFAULT NULL COMMENT '检测时间',
                              `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '状态(0:正常/1:禁用)',
                              `update_time` datetime DEFAULT NULL COMMENT '更新时间',
                              `create_time` datetime DEFAULT NULL COMMENT '创建时间',
                              PRIMARY KEY (`id`),
                              KEY `idx_createTime` (`create_time`),
                              KEY `idx_expireDay` (`expire_day`),
                              KEY `idx_runDay` (`run_day`),
                              KEY `idx_serverId` (`server_id`),
                              KEY `idx_serverId_uinStatus` (`server_id`,`uin_status`),
                              KEY `idx_serverKey` (`server_key`),
                              KEY `idx_status` (`status`),
                              KEY `idx_timing_runTime` (`timing`,`run_time`),
                              KEY `idx_uin` (`uin`),
                              KEY `idx_uinStatus` (`uin_status`),
                              KEY `idx_uinStatus_status_checkTime` (`uin_status`,`status`,`check_time`),
                              KEY `idx_updateTime` (`update_time`),
                              KEY `idx_userId` (`user_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;

CREATE TABLE `cloud_user_logs` (
                                   `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '日志id',
                                   `user_id` int(11) NOT NULL DEFAULT '0' COMMENT '用户ID',
                                   `request_ip` varchar(255) DEFAULT NULL COMMENT '操作IP',
                                   `type` tinyint(1) NOT NULL DEFAULT '0' COMMENT '事件类型(0:登陆/1:查看/2:删除/3:修改/4:创建)',
                                   `title` varchar(255) NOT NULL COMMENT '事件标题',
                                   `content` varchar(999) DEFAULT NULL COMMENT '事件详情',
                                   `update_time` datetime DEFAULT NULL COMMENT '更新时间',
                                   `create_time` datetime DEFAULT NULL COMMENT '创建时间',
                                   PRIMARY KEY (`id`),
                                   KEY `idx_requestIp` (`request_ip`),
                                   KEY `idx_userId_type` (`user_id`,`type`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;


CREATE TABLE `cloud_uin_notifies` (
                                      `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '事件ID',
                                      `user_id` int(11) NOT NULL DEFAULT '0' COMMENT '用户ID',
                                      `uin` int(11) DEFAULT NULL COMMENT '挂机账号',
                                      `type` int(11) DEFAULT NULL COMMENT '事件类型(0:状态异常/1:账号到期)',
                                      `status` tinyint(4) NOT NULL DEFAULT '0' COMMENT '状态(0:未读/1:已读)',
                                      `info` text NOT NULL COMMENT '内容',
                                      `update_time` datetime NOT NULL COMMENT '更新时间',
                                      `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                      PRIMARY KEY (`id`),
                                      KEY `idx_uin` (`uin`),
                                      KEY `idx_userId` (`user_id`),
                                      KEY `idx_userId_status_type` (`user_id`,`status`,`type`),
                                      KEY `idx_userId_uin` (`user_id`,`uin`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='QQ事件通知';



CREATE TABLE `cloud_users` (
                               `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '用户ID',
                               `phone` bigint(20) NOT NULL COMMENT '手机号',
                               `password` varchar(100) DEFAULT NULL COMMENT '用户密码',
                               `money` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '用户余额',
                               `login_token` varchar(255) DEFAULT NULL COMMENT '登陆token',
                               `login_ip` varchar(50) DEFAULT NULL COMMENT '登陆IP',
                               `login_time` datetime DEFAULT NULL COMMENT '登陆时间',
                               `status` int(1) NOT NULL DEFAULT '0' COMMENT '用户状态(0:正常/1:封禁)',
                               `create_time` datetime DEFAULT NULL COMMENT '创建时间',
                               `update_time` datetime DEFAULT NULL COMMENT '修改时间',
                               PRIMARY KEY (`id`),
                               KEY `idx_createTime` (`create_time`),
                               KEY `idx_status` (`status`),
                               KEY `idx_updateTime` (`update_time`),
                               KEY `idx_phone` (`phone`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COMMENT='网站用户表';


INSERT INTO `cloud_configs` (`id`, `ident_id`, `key`, `value`, `create_time`, `update_time`)
VALUES (1, 'Sys', 'web.name', '加速乐', '2023-11-10 10:13:24', '2024-10-12 17:56:34'),
       (2, 'Sys', 'web.title', '加速乐', '2023-11-10 10:13:49', '2024-10-12 17:56:34'),
       (3, 'Sys', 'web.keywords', '云端加速', '2023-11-10 10:14:08', '2024-10-12 17:56:34'),
       (4, 'Sys', 'web.describe', '云端加速', '2023-11-10 10:14:25', '2024-10-12 17:56:35'),
       (5, 'Sys', 'web.contact', '10001', '2023-11-10 10:14:40', '2024-10-12 17:56:34'),
       (6, 'Sys', 'ali.key', 'abcd', '2023-11-10 10:14:59', '2024-10-12 17:56:35'),
       (7, 'Sys', 'ali.secret', 'abcd', '2023-11-10 10:15:14', '2024-10-12 17:56:35'),
       (8, 'Sys', 'ali.sms.sign', '发信签名', '2023-11-10 10:15:32', '2024-10-12 17:56:35'),
       (9, 'Sys', 'ali.sms.template', '发信模板', '2023-11-10 10:15:46', '2024-10-12 17:56:35'),
       (10, 'Sys', 'ali.sms.var', '模板变量', '2023-11-10 10:16:02', '2024-10-12 17:56:36'),
       (11, 'Sys', 'uin.price.month', '3', '2024-05-06 16:20:00', '2024-10-12 17:56:36'),
       (12, 'Sys', 'uin.price.quarter', '10', '2024-05-06 16:20:00', '2024-10-12 17:56:36'),
       (13, 'Sys', 'uin.price.half', '18', '2024-05-06 16:20:00', '2024-10-12 17:56:36'),
       (14, 'Sys', 'uin.price.year', '30', '2024-05-06 16:20:00', '2024-10-12 17:56:36'),
       (15, 'Sys', 'uin.price.forever', '68', '2024-05-06 16:20:00', '2024-10-12 17:56:37'),
       (16, 'Sys', 'proto.qrcode', '2', '2024-09-28 22:59:13', '2024-10-12 17:56:37'),
       (17, 'Sys', 'proto.pwd', '2', '2024-09-28 22:59:39', '2024-10-12 17:56:37'),
       (18, 'Sys', 'proto.sign.android', 'http://127.0.0.1:8686', '2024-09-28 23:02:38', '2024-10-12 17:56:37'),
       (19, 'Sys', 'proto.sign.pc', 'http://127.0.0.1:8787', '2024-09-28 23:03:05', '2024-10-12 17:56:37'),
       (20, 'Sys', 'proto.auth.level', '100', '2024-10-14 20:52:53', '2024-10-25 13:59:00'),
       (21, 'Sys', 'web.url', 'http://jsl.dgjiabei.com', '2024-10-25 13:55:58', '2024-10-25 13:58:54'),
       (22, 'Sys', 'pay.min', '0.01', '2024-10-25 13:57:00', '2024-10-25 13:58:56'),
       (23, 'Sys', 'pay.max', '9999', '2024-10-25 13:57:01', '2024-10-25 13:58:56'),
       (24, 'Sys', 'pay.ali', '0', '2024-10-25 13:57:01', '2024-10-25 13:58:56'),
       (25, 'Sys', 'pay.wx', '0', '2024-10-25 13:57:01', '2024-10-25 13:58:56'),
       (26, 'Sys', 'pay.qq', '0', '2024-10-25 13:57:02', '2024-10-25 13:58:57'),
       (27, 'Sys', 'pay.pid', '1000', '2024-10-25 13:57:02', '2024-10-25 13:58:57'),
       (28, 'Sys', 'pay.key', '2Bp86XJ20jY9b93ybxx8wh8p3osImYJ5', '2024-10-25 13:57:03', '2024-10-25 13:58:57'),
       (29, 'Sys', 'pay.url', 'https://yzf.bgydg.cn/', '2024-10-25 13:57:03', '2024-10-25 13:58:57'),
       (30, 'Sys', 'refund.open', '1', '2024-10-25 13:57:32', '2024-10-25 13:58:58'),
       (31, 'Sys', 'refund.point', '3', '2024-10-25 13:57:32', '2024-10-25 13:58:59'),
       (32, 'Sys', 'refund.min', '0.1', '2024-10-25 13:57:33', '2024-10-25 13:58:59'),
       (33, 'Sys', 'refund.max', '30', '2024-10-25 13:57:33', '2024-10-25 13:58:59'),
       (34, 'Sys', 'refund.day', '30', '2024-10-25 13:57:33', '2024-10-25 13:58:59'),
       (35, 'Sys', 'refund.free', '3', '2024-10-25 13:57:34', '2024-10-25 13:58:59');

INSERT INTO `cloud_users` (`id`, `phone`, `password`, `money`, `login_token`, `login_ip`, `login_time`, `status`, `create_time`, `update_time`) VALUES (NULL, '18888888888', '$2a$10$xgvz5hRqn8BV7AN50cSLZeWkIS0vbxF3hXzmy8/turar7Le7sewpO', '10000.00', '{}', '127.0.0.1', '2024-10-12 23:08:03', '0', '2023-11-10 10:31:50', '2024-10-12 23:08:03')


