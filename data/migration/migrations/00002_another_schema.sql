-- +goose Up
-- +goose StatementBegin
INSERT INTO users (id, email, firstname, lastname, created_at)
VALUES (uuid_generate_v4(), 'bob.ross@example.com', 'Bob', 'Ross', CURRENT_TIMESTAMP),
    (uuid_generate_v4(), 'jimmy.page@example.com', 'Jimmy', 'Page', CURRENT_TIMESTAMP);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM users
WHERE (email, firstname, lastname) IN (
    ('bob.ross@example.com', 'Bob', 'Ross'),
    ('jimmy.page@example.com', 'Jimmy', 'Page')
);
-- +goose StatementEnd
