-- Insert test company
INSERT OR IGNORE INTO companies (cnpj, name, email, fone, active, created_at, updated_at) 
VALUES ('12345678000195', 'Test Company', 'test@company.com', '(11) 99999-9999', 1, datetime('now'), datetime('now'));

-- Insert test admin user
INSERT OR IGNORE INTO employees (name, cpf, rg, email, age, active, workload, is_manager, is_admin, company_cnpj, created_at, updated_at) 
VALUES ('Admin User', '12345678901', '123456789', 'admin@test.com', 30, 1, 40.0, 0, 1, '12345678000195', datetime('now'), datetime('now'));

-- Insert test manager
INSERT OR IGNORE INTO employees (name, cpf, rg, email, age, active, workload, is_manager, is_admin, company_cnpj, created_at, updated_at) 
VALUES ('Manager User', '98765432109', '987654321', 'manager@test.com', 35, 1, 40.0, 1, 0, '12345678000195', datetime('now'), datetime('now'));

-- Insert test employee
INSERT OR IGNORE INTO employees (name, cpf, rg, email, age, active, workload, is_manager, is_admin, company_cnpj, created_at, updated_at) 
VALUES ('Employee User', '11122233344', '111222333', 'employee@test.com', 25, 1, 40.0, 0, 0, '12345678000195', datetime('now'), datetime('now'));

-- Insert login credentials (password is 'test123!')
INSERT OR IGNORE INTO logins (email, password, created_at, updated_at) 
VALUES ('admin@test.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', datetime('now'), datetime('now'));

INSERT OR IGNORE INTO logins (email, password, created_at, updated_at) 
VALUES ('manager@test.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', datetime('now'), datetime('now'));

INSERT OR IGNORE INTO logins (email, password, created_at, updated_at) 
VALUES ('employee@test.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', datetime('now'), datetime('now'));
