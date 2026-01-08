-- ============================================
-- PT BESQ Enhanced Database Schema v2.0
-- Compatible with MariaDB 10.6
-- ============================================

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

-- Create database
CREATE DATABASE IF NOT EXISTS `besq_db` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `besq_db`;

-- ============================================
-- 1. USERS TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS `users` (
  `id` INT AUTO_INCREMENT PRIMARY KEY,
  `username` VARCHAR(50) UNIQUE NOT NULL,
  `password_hash` VARCHAR(255) NOT NULL,
  `role` ENUM('admin', 'operator', 'supervisor', 'viewer') DEFAULT 'operator',
  `email` VARCHAR(100) UNIQUE,
  `full_name` VARCHAR(100),
  `phone` VARCHAR(20),
  `is_active` TINYINT(1) DEFAULT 1,
  `last_login` TIMESTAMP NULL,
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_username (username),
  INDEX idx_role (role),
  INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- 2. ACTIVITY LOGS TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS `activity_logs` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `user_id` INT,
  `username` VARCHAR(50),
  `method` VARCHAR(10),
  `path` VARCHAR(255),
  `ip_address` VARCHAR(45),
  `status_code` INT,
  `user_agent` TEXT,
  `request_body` TEXT,
  `response_time_ms` INT,
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_user_id (user_id),
  INDEX idx_created_at (created_at),
  INDEX idx_path (path),
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- 3. PROCESS TEMPLATES
-- ============================================
CREATE TABLE IF NOT EXISTS `process_templates` (
  `id` INT AUTO_INCREMENT PRIMARY KEY,
  `name` VARCHAR(100) NOT NULL,
  `description` TEXT,
  `category` VARCHAR(50),
  `icon` VARCHAR(50),
  `color` VARCHAR(20),
  `estimated_duration` INT COMMENT 'in minutes',
  `is_active` TINYINT(1) DEFAULT 1,
  `created_by` INT,
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_category (category),
  INDEX idx_is_active (is_active),
  FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- 4. FIELD DEFINITIONS
-- ============================================
CREATE TABLE IF NOT EXISTS `field_definitions` (
  `id` INT AUTO_INCREMENT PRIMARY KEY,
  `template_id` INT NOT NULL,
  `field_key` VARCHAR(50) NOT NULL,
  `field_label` VARCHAR(100) NOT NULL,
  `field_type` ENUM('text', 'number', 'date', 'datetime', 'select', 'textarea', 'checkbox', 'file') NOT NULL,
  `is_required` TINYINT(1) DEFAULT 0,
  `validation_rule` TEXT COMMENT 'JSON format',
  `default_value` VARCHAR(255),
  `placeholder` VARCHAR(255),
  `help_text` TEXT,
  `display_order` INT DEFAULT 0,
  `is_active` TINYINT(1) DEFAULT 1,
  INDEX idx_template_id (template_id),
  INDEX idx_display_order (display_order),
  FOREIGN KEY (template_id) REFERENCES process_templates(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- 5. WORKFLOWS
-- ============================================
CREATE TABLE IF NOT EXISTS `workflows` (
  `id` INT AUTO_INCREMENT PRIMARY KEY,
  `name` VARCHAR(100) NOT NULL,
  `description` TEXT,
  `canvas_config` LONGTEXT,
  `version` INT DEFAULT 1,
  `is_active` TINYINT(1) DEFAULT 1,
  `is_published` TINYINT(1) DEFAULT 0,
  `created_by` INT,
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_is_active (is_active),
  INDEX idx_created_by (created_by),
  FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- 6. PROCESS INSTANCES
-- ============================================
CREATE TABLE IF NOT EXISTS `process_instances` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `template_id` INT NOT NULL,
  `workflow_id` INT NOT NULL,
  `batch_number` VARCHAR(50) UNIQUE,
  `data_payload` LONGTEXT,
  `status` ENUM('draft', 'in_progress', 'completed', 'rejected', 'cancelled') DEFAULT 'draft',
  `priority` ENUM('low', 'normal', 'high', 'urgent') DEFAULT 'normal',
  `start_time` TIMESTAMP NULL,
  `end_time` TIMESTAMP NULL,
  `duration_minutes` INT,
  `notes` TEXT,
  `created_by` INT,
  `approved_by` INT,
  `approved_at` TIMESTAMP NULL,
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_template_id (template_id),
  INDEX idx_workflow_id (workflow_id),
  INDEX idx_status (status),
  INDEX idx_batch_number (batch_number),
  INDEX idx_created_at (created_at),
  INDEX idx_created_by (created_by),
  FOREIGN KEY (template_id) REFERENCES process_templates(id),
  FOREIGN KEY (workflow_id) REFERENCES workflows(id),
  FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
  FOREIGN KEY (approved_by) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- 7. INSTANCE HISTORY
-- ============================================
CREATE TABLE IF NOT EXISTS `instance_history` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `instance_id` BIGINT NOT NULL,
  `action` ENUM('created', 'updated', 'status_changed', 'approved', 'rejected') NOT NULL,
  `old_value` TEXT,
  `new_value` TEXT,
  `changed_by` INT,
  `comment` TEXT,
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_instance_id (instance_id),
  INDEX idx_action (action),
  INDEX idx_created_at (created_at),
  FOREIGN KEY (instance_id) REFERENCES process_instances(id) ON DELETE CASCADE,
  FOREIGN KEY (changed_by) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- 8. NOTIFICATIONS
-- ============================================
CREATE TABLE IF NOT EXISTS `notifications` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `user_id` INT NOT NULL,
  `type` ENUM('info', 'warning', 'error', 'success') DEFAULT 'info',
  `title` VARCHAR(255) NOT NULL,
  `message` TEXT,
  `related_entity_type` VARCHAR(50) COMMENT 'instance, workflow, user',
  `related_entity_id` BIGINT,
  `is_read` TINYINT(1) DEFAULT 0,
  `read_at` TIMESTAMP NULL,
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_user_id (user_id),
  INDEX idx_is_read (is_read),
  INDEX idx_created_at (created_at),
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- 9. SYSTEM SETTINGS
-- ============================================
CREATE TABLE IF NOT EXISTS `system_settings` (
  `id` INT AUTO_INCREMENT PRIMARY KEY,
  `setting_key` VARCHAR(100) UNIQUE NOT NULL,
  `setting_value` TEXT,
  `setting_type` ENUM('string', 'number', 'boolean', 'json') DEFAULT 'string',
  `description` TEXT,
  `is_public` TINYINT(1) DEFAULT 0 COMMENT 'Can be accessed without auth',
  `updated_by` INT,
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_setting_key (setting_key),
  FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- 10. FILE ATTACHMENTS
-- ============================================
CREATE TABLE IF NOT EXISTS `file_attachments` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `instance_id` BIGINT,
  `filename` VARCHAR(255) NOT NULL,
  `original_filename` VARCHAR(255),
  `file_path` VARCHAR(500),
  `file_size` BIGINT COMMENT 'in bytes',
  `mime_type` VARCHAR(100),
  `uploaded_by` INT,
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_instance_id (instance_id),
  INDEX idx_uploaded_by (uploaded_by),
  FOREIGN KEY (instance_id) REFERENCES process_instances(id) ON DELETE CASCADE,
  FOREIGN KEY (uploaded_by) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- 11. SCHEDULED REPORTS
-- ============================================
CREATE TABLE IF NOT EXISTS `scheduled_reports` (
  `id` INT AUTO_INCREMENT PRIMARY KEY,
  `name` VARCHAR(100) NOT NULL,
  `report_type` VARCHAR(50),
  `schedule` VARCHAR(50) COMMENT 'daily, weekly, monthly',
  `recipients` TEXT COMMENT 'JSON array of emails',
  `filters` TEXT COMMENT 'JSON filters',
  `is_active` TINYINT(1) DEFAULT 1,
  `last_run` TIMESTAMP NULL,
  `next_run` TIMESTAMP NULL,
  `created_by` INT,
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_is_active (is_active),
  INDEX idx_next_run (next_run),
  FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- INSERT SAMPLE DATA
-- ============================================

-- Insert default admin user (password: admin123)
INSERT INTO users (username, password_hash, role, email, full_name, is_active) VALUES
('admin', '$2a$14$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5/qLYzW.6YQ2i', 'admin', 'admin@besq.com', 'System Administrator', 1)
ON DUPLICATE KEY UPDATE username=username;

-- Insert sample template
INSERT INTO process_templates (name, description, category, icon, color, estimated_duration, is_active, created_by) VALUES
('Mixing', 'Rubber mixing process', 'Production', 'mix', '#3B82F6', 60, 1, 1),
('Oven Curing', 'High temperature curing', 'Production', 'fire', '#EF4444', 120, 1, 1)
ON DUPLICATE KEY UPDATE name=name;

-- Insert sample workflow
INSERT INTO workflows (name, description, canvas_config, is_active, created_by) VALUES
('Production Line A', 'Main production workflow', '{"nodes": [], "edges": []}', 1, 1)
ON DUPLICATE KEY UPDATE name=name;

-- Insert field definitions for Mixing template
INSERT INTO field_definitions (template_id, field_key, field_label, field_type, is_required, placeholder, display_order) VALUES
(1, 'batch_code', 'Batch Code', 'text', 1, 'e.g., BATCH-001', 1),
(1, 'rubber_weight', 'Rubber Weight (kg)', 'number', 1, 'Enter weight', 2),
(1, 'temperature', 'Temperature (°C)', 'number', 1, 'Enter temperature', 3),
(1, 'operator_notes', 'Operator Notes', 'textarea', 0, 'Optional notes', 4)
ON DUPLICATE KEY UPDATE field_key=field_key;

-- Insert field definitions for Oven Curing template
INSERT INTO field_definitions (template_id, field_key, field_label, field_type, is_required, placeholder, display_order) VALUES
(2, 'batch_code', 'Batch Code', 'text', 1, 'e.g., BATCH-001', 1),
(2, 'oven_temp', 'Oven Temperature (°C)', 'number', 1, 'Enter temperature', 2),
(2, 'curing_time', 'Curing Time (minutes)', 'number', 1, 'Enter time', 3),
(2, 'pressure', 'Pressure (bar)', 'number', 1, 'Enter pressure', 4)
ON DUPLICATE KEY UPDATE field_key=field_key;

-- Insert system settings
INSERT INTO system_settings (setting_key, setting_value, setting_type, description, is_public, updated_by) VALUES
('company_name', 'PT Besq Manufacturing', 'string', 'Company name displayed in the system', 1, 1),
('max_file_upload_size', '10485760', 'number', 'Maximum file upload size in bytes (10MB)', 0, 1),
('session_timeout_minutes', '1440', 'number', 'User session timeout in minutes', 0, 1),
('enable_notifications', 'true', 'boolean', 'Enable real-time notifications', 0, 1)
ON DUPLICATE KEY UPDATE setting_key=setting_key;

-- Insert sample process instance
INSERT INTO process_instances (template_id, workflow_id, batch_number, data_payload, status, priority, created_by) VALUES
(1, 1, 'BATCH-001', '{"batch_code":"BATCH-001","rubber_weight":50.5,"temperature":180,"operator_notes":"All parameters normal"}', 'completed', 'normal', 1)
ON DUPLICATE KEY UPDATE batch_number=batch_number;

-- ============================================
-- CLEANUP
-- ============================================
/*!40103 SET TIME_ZONE=IFNULL(@OLD_TIME_ZONE, 'system') */;
/*!40101 SET SQL_MODE=IFNULL(@OLD_SQL_MODE, '') */;
/*!40014 SET FOREIGN_KEY_CHECKS=IFNULL(@OLD_FOREIGN_KEY_CHECKS, 1) */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40111 SET SQL_NOTES=IFNULL(@OLD_SQL_NOTES, 1) */;