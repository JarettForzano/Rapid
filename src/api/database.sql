CREATE TABLE IF NOT EXISTS transfer 
(
    id SERIAL PRIMARY KEY, 
    from_user INTEGER NOT NULL, 
    to_user INTEGER NOT NULL, 
    filename TEXT NOT NULL
);


CREATE TABLE IF NOT EXISTS users 
(
    id SERIAL PRIMARY KEY, 
    username VARCHAR(100) NOT NULL,
    nickname VARCHAR(100) NOT NULL,
    password VARCHAR(100) NOT NULL,
    friend_code VARCHAR(100), 
    uuid VARCHAR(100),
    session INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS userkey
(
    users_id INTEGER NOT NULL,
    key VARCHAR(1000)
);

CREATE TABLE IF NOT EXISTS friends
(
    id SERIAL PRIMARY KEY,
    user_one INTEGER NOT NULL, 
    user_two INTEGER NOT NULL
);