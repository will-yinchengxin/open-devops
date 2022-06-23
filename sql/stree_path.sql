set names utf8;
drop database if exists open_devops;
CREATE DATABASE IF NOT EXISTS  open_devops charset utf8 COLLATE utf8_general_ci;

--------------------------------------------------------------------------------------------
drop table stree_path;

CREATE TABLE `stree_path` (
      `id` int(11) NOT NULL AUTO_INCREMENT,
      `level` tinyint(4) NOT NULL,
      `path` varchar(200) DEFAULT NULL,
      `node_name` varchar(200) DEFAULT NULL,
      PRIMARY KEY (`id`),
      UNIQUE KEY `idx_unique_key` (`level`,`path`,`node_name`) USING BTREE COMMENT '唯一索引'
) ENGINE=InnoDB  DEFAULT CHARSET=utf8;

--------------------------------------------------------------------------------------------
drop table resource_host_test;

CREATE TABLE `resource_host_test` (
      `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增id',
      `name` varchar(200) NOT NULL COMMENT '资源名称',
      `tags` varchar(1024)  DEFAULT ''  COMMENT '标签map',
      `private_ips` varchar(1024)  DEFAULT ''  COMMENT '内网IP数组',
      PRIMARY KEY (`id`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4;

--------------------------------------------------------------------------------------------
drop table resource_host;

CREATE TABLE `resource_host` (

             `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增id',
             `uid` varchar(100) NOT NULL COMMENT '实例id',
             `hash` varchar(100) NOT NULL COMMENT '哈希',
             `name` varchar(200) NOT NULL COMMENT '资源名称',
             `private_ips` varchar(1024)  DEFAULT ''  COMMENT '内网IP数组',
             `tags` varchar(1024)  DEFAULT ''  COMMENT '标签map',
    -- 公有云字段
             `cloud_provider` varchar(20) NOT NULL COMMENT '云类型',
             `charging_mode` varchar(10) DEFAULT NULL COMMENT '付费类型',
             `region` varchar(20) NOT NULL COMMENT '标签region',
             `account_id` int(11) NOT NULL COMMENT '对应账户在account表中的id',
             `vpc_id` varchar(40) DEFAULT NULL COMMENT 'VPC ID',
             `subnet_id` varchar(40) DEFAULT NULL COMMENT '子网ID',
             `security_groups`  varchar(1024)  DEFAULT '' COMMENT '安全组',
             `status` varchar(20) NOT NULL COMMENT '状态',
             `instance_type` varchar(100) NOT NULL COMMENT '资产规格类型',
             `public_ips` varchar(1024)  DEFAULT ''  COMMENT '公网网IP数组',
             `availability_zone` varchar(20) NOT NULL COMMENT '可用区',
    -- 机器字段
             `cpu` varchar(20) NOT NULL COMMENT 'cpu核数',
             `mem` varchar(20) NOT NULL COMMENT '内存g数',
             `disk` varchar(20) NOT NULL COMMENT '磁盘g数',
             `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
             `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    -- 服务树字段(新增, 开篇部分提及过)
             `stree_group` varchar(100) NOT NULL COMMENT '服务树g字段',
             `stree_product` varchar(100) NOT NULL COMMENT '服务树p字段',
             `stree_app` varchar(100) NOT NULL COMMENT '服务树a字段',

             PRIMARY KEY (`id`),
             UNIQUE KEY `hash_uid` (`uid`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4 ;
