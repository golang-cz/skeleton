-- +goose Up
-- +goose StatementBegin
INSERT INTO users (email, firstname, lastname, created_at)
VALUES ('bob.ross@example.com', 'Bob', 'Ross', CURRENT_TIMESTAMP),
    ('jimmy.page@example.com', 'Jimmy', 'Page', CURRENT_TIMESTAMP);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM users
WHERE email IN ('bob.ross@example.com', 'jimmy.page@example.com');
-- +goose StatementEnd
