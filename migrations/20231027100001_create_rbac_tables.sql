-- +goose Up
-- +goose StatementBegin
-- Buat tabel 'roles'
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Buat tabel 'permissions'
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    name VARCHAR(100) UNIQUE NOT NULL, -- contoh: "users:create", "dashboard:view"
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Buat tabel join 'role_permissions' (many-to-many)
CREATE TABLE role_permissions (
    role_id UUID NOT NULL,
    permission_id UUID NOT NULL,
    PRIMARY KEY (role_id, permission_id),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);

-- Tambahkan kolom 'role_id' ke tabel 'users' dan foreign key
ALTER TABLE users ADD COLUMN role_id UUID;
ALTER TABLE users ADD FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE SET NULL;

-- Hapus constraint 'CHECK' lama pada kolom 'role' (sudah tidak relevan)
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;

-- Buat index untuk foreign key (opsional, tapi baik untuk performa)
CREATE INDEX idx_users_role_id ON users(role_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Hapus index
DROP INDEX IF EXISTS idx_users_role_id;

-- Kembalikan constraint 'CHECK' lama (jika perlu, tergantung apakah Anda masih memerlukannya)
-- ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('admin', 'user'));

-- Hapus foreign key dan kolom 'role_id' dari tabel 'users'
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_id_fkey;
ALTER TABLE users DROP COLUMN IF EXISTS role_id;

-- Hapus tabel join 'role_permissions'
DROP TABLE IF EXISTS role_permissions;

-- Hapus tabel 'permissions'
DROP TABLE IF EXISTS permissions;

-- Hapus tabel 'roles'
DROP TABLE IF EXISTS roles;

-- +goose StatementEnd