-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION "uuid-ossp";

INSERT INTO users (id, email, firstname, lastname, created_at)
VALUES (uuid_generate_v4(), 'bob.ross@happy-little-accident.com', 'Bob', 'Ross', CURRENT_TIMESTAMP),
    (uuid_generate_v4(), 'jimmy.page@yardbirds.com', 'Jimmy', 'Page', CURRENT_TIMESTAMP);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP EXTENSION IF EXISTS "uuid-ossp";
DELETE FROM users
WHERE (email, firstname, lastname) IN (
    ('bob.ross@happy-little-accident.com', 'Bob', 'Ross'),
    ('jimmy.page@yardbirds.com', 'Jimmy', 'Page')
);
-- +goose StatementEnd
