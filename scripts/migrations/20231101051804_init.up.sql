CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE users
(
    id                SERIAL PRIMARY KEY,
    email             VARCHAR(255) UNIQUE NOT NULL,
    password_hash     VARCHAR(255),
    created_at        TIMESTAMP           NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP           NOT NULL DEFAULT CURRENT_TIMESTAMP,
    email_verified_at TIMESTAMP,
    deleted_at        TIMESTAMP
);

CREATE TABLE otps
(
    id         SERIAL PRIMARY KEY,
    user_id    INTEGER    NOT NULL REFERENCES users (id),
    otp_code   VARCHAR(6) NOT NULL,
    is_used    BOOLEAN    NOT NULL DEFAULT FALSE,
    expires_at TIMESTAMP  NOT NULL,
    created_at TIMESTAMP  NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE contacts
(
    id                 SERIAL PRIMARY KEY,
    name               VARCHAR(255),
    avatar             VARCHAR(255),
    activity_name      TEXT,
    about              TEXT,
    website            VARCHAR(255),
    country_code       VARCHAR(10),
    views_amount       INTEGER            DEFAULT 0,
    saves_amount       INTEGER            DEFAULT 0,
    created_at         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at         TIMESTAMP,
    phone_number       VARCHAR(50),
    phone_calling_code VARCHAR(10),
    email              VARCHAR(255),
    user_id      INTEGER NOT NULL REFERENCES users (id),
    is_published BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE addresses
(
    id          SERIAL PRIMARY KEY,
    external_id VARCHAR(255),
    contact_id  INTEGER   NOT NULL REFERENCES contacts (id),
    label       VARCHAR(255),
    name        VARCHAR(255),
    location    GEOGRAPHY(POINT, 4326),
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at  TIMESTAMP
);

CREATE TABLE tags
(
    id   SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL
);

CREATE TABLE contact_tags
(
    contact_id INTEGER NOT NULL REFERENCES contacts (id),
    tag_id     INTEGER NOT NULL REFERENCES tags (id),
    PRIMARY KEY (contact_id, tag_id)
);

CREATE TABLE social_media_links
(
    id         SERIAL PRIMARY KEY,
    contact_id INTEGER NOT NULL REFERENCES contacts (id),
    type       VARCHAR(50),
    link       VARCHAR(255),
    label      VARCHAR(255)
);

CREATE TABLE saved_contacts
(
    user_id    INTEGER NOT NULL REFERENCES users (id),
    contact_id INTEGER NOT NULL REFERENCES contacts (id),
    PRIMARY KEY (user_id, contact_id)
);

-- update number of saves when new row is added to saved_contacts or deleted

CREATE OR REPLACE FUNCTION increment_saves()
    RETURNS TRIGGER AS
$$
BEGIN
    UPDATE contacts
    SET saves_amount = saves_amount + 1
    WHERE id = NEW.contact_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_increment_saves
    AFTER INSERT
    ON saved_contacts
    FOR EACH ROW
EXECUTE FUNCTION increment_saves();

CREATE OR REPLACE FUNCTION decrement_saves()
    RETURNS TRIGGER AS
$$
BEGIN
    UPDATE contacts
    SET saves_amount = saves_amount - 1
    WHERE id = OLD.contact_id;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_decrement_saves
    AFTER DELETE
    ON saved_contacts
    FOR EACH ROW
EXECUTE FUNCTION decrement_saves();
