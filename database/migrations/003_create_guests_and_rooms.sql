CREATE TABLE guests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(120) NOT NULL,
    document VARCHAR(30),
    email VARCHAR(255),
    phone VARCHAR(30),
    address TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT guests_name_not_blank CHECK (length(trim(name)) > 0)
);

CREATE UNIQUE INDEX guests_document_unique_idx ON guests (document) WHERE document IS NOT NULL;
CREATE UNIQUE INDEX guests_email_unique_idx ON guests (lower(email)) WHERE email IS NOT NULL;
CREATE INDEX guests_name_idx ON guests (name);

CREATE TABLE room_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(80) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT room_categories_name_unique UNIQUE (name),
    CONSTRAINT room_categories_name_not_blank CHECK (length(trim(name)) > 0)
);

CREATE TABLE rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    number VARCHAR(20) NOT NULL,
    category_id UUID NOT NULL REFERENCES room_categories(id),
    capacity INTEGER NOT NULL CHECK (capacity >= 1),
    daily_rate NUMERIC(12, 2) NOT NULL CHECK (daily_rate >= 0),
    is_available BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT rooms_number_unique UNIQUE (number)
);

CREATE INDEX rooms_category_id_idx ON rooms (category_id);
CREATE INDEX rooms_availability_idx ON rooms (is_available);
