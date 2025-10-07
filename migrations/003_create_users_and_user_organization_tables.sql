-- +goose Up
-- +goose StatementBegin

-- Create users table with multi-authentication support
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE,
    password VARCHAR(255), -- Nullable for OAuth users
    role_id UUID REFERENCES roles(id) ON DELETE SET NULL,
    google_id VARCHAR(255), -- Will have partial unique constraint
    avatar_url TEXT,
    auth_provider VARCHAR(20) DEFAULT 'local', -- Track authentication method
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Create partial unique constraint for google_id (best practice for nullable unique fields)
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_google_id_unique 
ON users(google_id) 
WHERE google_id IS NOT NULL;

-- Add check constraint for auth provider consistency
ALTER TABLE users ADD CONSTRAINT chk_auth_provider_consistency 
CHECK (
    (auth_provider = 'google' AND google_id IS NOT NULL) OR
    (auth_provider = 'local' AND google_id IS NULL) OR
    (auth_provider NOT IN ('google', 'local'))
);

-- Create user_organizations junction table with role assignment
CREATE TABLE IF NOT EXISTS user_organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    role_id UUID REFERENCES roles(id) ON DELETE SET NULL,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    left_at TIMESTAMPTZ,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE(user_id, organization_id)
);

-- Create user_organization_history for audit trail
CREATE TABLE IF NOT EXISTS user_organization_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    user_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    old_role_id UUID,
    new_role_id UUID,
    changed_by UUID REFERENCES users(id),
    change_reason TEXT,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create role_approvals table for role request workflow
CREATE TABLE IF NOT EXISTS role_approvals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    requested_role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    approver_id UUID REFERENCES users(id),
    status VARCHAR(20) DEFAULT 'pending',
    request_message TEXT,
    approval_message TEXT,
    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Create indexes for users
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role_id ON users(role_id);
CREATE INDEX IF NOT EXISTS idx_users_auth_provider ON users(auth_provider);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

-- Create indexes for user_organizations
CREATE INDEX IF NOT EXISTS idx_user_orgs_user_id ON user_organizations(user_id);
CREATE INDEX IF NOT EXISTS idx_user_orgs_org_id ON user_organizations(organization_id);
CREATE INDEX IF NOT EXISTS idx_user_orgs_role_id ON user_organizations(role_id);
CREATE INDEX IF NOT EXISTS idx_user_orgs_is_active ON user_organizations(is_active);
CREATE INDEX IF NOT EXISTS idx_user_orgs_deleted_at ON user_organizations(deleted_at);

-- Create indexes for user_organization_history
CREATE INDEX IF NOT EXISTS idx_user_org_history_user_id ON user_organization_history(user_id);
CREATE INDEX IF NOT EXISTS idx_user_org_history_org_id ON user_organization_history(organization_id);
CREATE INDEX IF NOT EXISTS idx_user_org_history_changed_by ON user_organization_history(changed_by);
CREATE INDEX IF NOT EXISTS idx_user_org_history_changed_at ON user_organization_history(changed_at);

-- Create indexes for role_approvals
CREATE INDEX IF NOT EXISTS idx_role_approvals_user_id ON role_approvals(user_id);
CREATE INDEX IF NOT EXISTS idx_role_approvals_org_id ON role_approvals(organization_id);
CREATE INDEX IF NOT EXISTS idx_role_approvals_role_id ON role_approvals(requested_role_id);
CREATE INDEX IF NOT EXISTS idx_role_approvals_approver_id ON role_approvals(approver_id);
CREATE INDEX IF NOT EXISTS idx_role_approvals_status ON role_approvals(status);
CREATE INDEX IF NOT EXISTS idx_role_approvals_deleted_at ON role_approvals(deleted_at);

-- Add foreign key constraint for organization creator after users table is created
ALTER TABLE organizations ADD CONSTRAINT fk_organizations_created_by 
FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove foreign key constraint first
ALTER TABLE organizations DROP CONSTRAINT IF EXISTS fk_organizations_created_by;

DROP TABLE IF EXISTS role_approvals;
DROP TABLE IF EXISTS user_organization_history;
DROP TABLE IF EXISTS user_organizations;
DROP TABLE IF EXISTS users;

-- +goose StatementEnd