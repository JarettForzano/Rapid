CREATE TABLE IF NOT EXISTS transfer 
(
    id SERIAL PRIMARY KEY, 
    from_user INTEGER NOT NULL, 
    to_user INTEGER NOT NULL, 
    key TEXT NOT NULL, 
    filename TEXT NOT NULL
);


CREATE TABLE IF NOT EXISTS users 
(
    id SERIAL PRIMARY KEY, 
    nickname VARCHAR(10) NOT NULL,
    username VARCHAR(10) NOT NULL,
    password VARCHAR(10) NOT NULL,
    friend_code VARCHAR(10), 
    uuid VARCHAR(100),
    session INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS userkey
(
    users_id INTEGER NOT NULL,
    key VARCHAR(MAX)
);

CREATE TABLE IF NOT EXISTS friends
(
    id SERIAL PRIMARY KEY,
    user_one INTEGER NOT NULL, 
    user_two INTEGER NOT NULL
);