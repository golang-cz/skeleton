-- +goose Up
-- +goose StatementBegin
CREATE TABLE users
(
    id          UUID PRIMARY KEY NOT NULL,
    email       VARCHAR(255) NOT NULL,
    firstname   VARCHAR(255) NOT NULL,
    lastname    VARCHAR(255) NOT NULL,
    created_at  TIMESTAMP    NOT NULL,
    updated_at  TIMESTAMP    NOT NULL,
    deleted_at  TIMESTAMP
);

CREATE INDEX users_id_idx ON users USING btree (id);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd

