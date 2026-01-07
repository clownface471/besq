-- --------------------------------------------------------
-- Host:                         127.0.0.1
-- Server version:               12.1.2-MariaDB - MariaDB Server
-- Server OS:                    Win64
-- HeidiSQL Version:             12.14.0.7165
-- --------------------------------------------------------

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET NAMES utf8 */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;


-- Dumping database structure for besq_db
CREATE DATABASE IF NOT EXISTS `besq_db` /*!40100 DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_uca1400_ai_ci */;
USE `besq_db`;

-- Dumping structure for table besq_db.field_definitions
CREATE TABLE IF NOT EXISTS `field_definitions` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `template_id` int(11) NOT NULL,
  `field_key` varchar(50) NOT NULL,
  `field_label` varchar(100) NOT NULL,
  `field_type` varchar(20) NOT NULL,
  `is_required` tinyint(1) DEFAULT 0,
  `validation_rule` text DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `template_id` (`template_id`),
  CONSTRAINT `1` FOREIGN KEY (`template_id`) REFERENCES `process_templates` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_ai_ci;

-- Dumping data for table besq_db.field_definitions: ~0 rows (approximately)

-- Dumping structure for table besq_db.process_instances
CREATE TABLE IF NOT EXISTS `process_instances` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `template_id` int(11) NOT NULL,
  `workflow_id` int(11) NOT NULL,
  `data_payload` longtext DEFAULT NULL CHECK (json_valid(`data_payload`)),
  `status` varchar(50) DEFAULT 'draft',
  `created_by` int(11) DEFAULT 0,
  `created_at` timestamp NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  PRIMARY KEY (`id`),
  KEY `idx_instance_template` (`template_id`),
  KEY `idx_instance_workflow` (`workflow_id`),
  KEY `idx_instance_created` (`created_at`),
  CONSTRAINT `1` FOREIGN KEY (`template_id`) REFERENCES `process_templates` (`id`),
  CONSTRAINT `2` FOREIGN KEY (`workflow_id`) REFERENCES `workflows` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_ai_ci;

-- Dumping data for table besq_db.process_instances: ~0 rows (approximately)
INSERT INTO `process_instances` (`id`, `template_id`, `workflow_id`, `data_payload`, `status`, `created_by`, `created_at`, `updated_at`) VALUES
	(1, 1, 1, '{"batch_code":"BATCH-001","rubber_weight":50.5}', 'draft', 0, '2026-01-07 05:28:57', '2026-01-07 05:28:57'),
	(2, 1, 1, '{"batch_code":"BATCH-001","rubber_weight":50.5}', 'draft', 0, '2026-01-07 05:40:23', '2026-01-07 05:40:23'),
	(3, 1, 1, '{"batch_code":"OP-TEST-001","rubber_weight":75.5}', 'draft', 0, '2026-01-07 06:21:09', '2026-01-07 06:21:09');

-- Dumping structure for table besq_db.process_templates
CREATE TABLE IF NOT EXISTS `process_templates` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL,
  `description` text DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT current_timestamp(),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_ai_ci;

-- Dumping data for table besq_db.process_templates: ~0 rows (approximately)
INSERT INTO `process_templates` (`id`, `name`, `description`, `created_at`) VALUES
	(1, 'Oven Curing', 'Pemanasan dalam suhu tinggi', '2026-01-07 04:22:30');

-- Dumping structure for table besq_db.workflows
CREATE TABLE IF NOT EXISTS `workflows` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL,
  `canvas_config` longtext DEFAULT NULL CHECK (json_valid(`canvas_config`)),
  `is_active` tinyint(1) DEFAULT 1,
  `created_at` timestamp NULL DEFAULT current_timestamp(),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_ai_ci;

-- Dumping data for table besq_db.workflows: ~0 rows (approximately)
INSERT INTO `workflows` (`id`, `name`, `canvas_config`, `is_active`, `created_at`) VALUES
	(1, 'Produksi Baut Standard', '{"nodes": [], "edges": []}', 1, '2026-01-07 04:48:33');

/*!40103 SET TIME_ZONE=IFNULL(@OLD_TIME_ZONE, 'system') */;
/*!40101 SET SQL_MODE=IFNULL(@OLD_SQL_MODE, '') */;
/*!40014 SET FOREIGN_KEY_CHECKS=IFNULL(@OLD_FOREIGN_KEY_CHECKS, 1) */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40111 SET SQL_NOTES=IFNULL(@OLD_SQL_NOTES, 1) */;
