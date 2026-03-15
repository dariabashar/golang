-- Очистка таблиц (опционально, для повторного запуска seed)
TRUNCATE TABLE user_friends RESTART IDENTITY CASCADE;
TRUNCATE TABLE users RESTART IDENTITY CASCADE;

INSERT INTO users (name, email, gender, birth_date) VALUES
('Дашуля Крутая',               'dashulya.krutaya@example.com',      'female', '2004-01-10'),
('Бакытжан Лучший_Преподаватель','bakytzhan.teacher@example.com',    'male',   '2003-03-22'),
('Alice',        'alice@example.com',        'female', '1995-01-10'),
('Bob',          'bob@example.com',          'male',   '1993-03-22'),
('Carol',        'carol@example.com',        'female', '1994-05-15'),
('Dave',         'dave@example.com',         'male',   '1992-07-30'),
('Eve',          'eve@example.com',          'female', '1996-11-05'),
('Frank',        'frank@example.com',        'male',   '1991-09-12'),
('Grace',        'grace@example.com',        'female', '1997-02-18'),
('Heidi',        'heidi@example.com',        'female', '1990-12-01'),
('Ivan',         'ivan@example.com',         'male',   '1993-08-08'),
('Judy',         'judy@example.com',         'female', '1994-04-04'),
('Ken',          'ken@example.com',          'male',   '1992-02-02'),
('Laura',        'laura@example.com',        'female', '1995-06-06'),
('Mallory',      'mallory@example.com',      'female', '1991-01-20'),
('Niaj',         'niaj@example.com',         'male',   '1990-03-03'),
('Olivia',       'olivia@example.com',       'female', '1998-09-09'),
('Peggy',        'peggy@example.com',        'female', '1997-10-10'),
('Rupert',       'rupert@example.com',       'male',   '1989-05-25'),
('Sybil',        'sybil@example.com',        'female', '1992-12-12');

-- IDs после вставки будут 1..20 в том же порядке.
-- Настроим дружбы.
-- Обеспечим, чтобы у пользователей 1 (Alice) и 2 (Bob) было >= 3 общих друга.
-- Общие друзья: 3 (Carol), 4 (Dave), 5 (Eve), 6 (Frank)

-- Alice's friends
INSERT INTO user_friends (user_id, friend_id) VALUES
-- общие с Bob
(1, 3),
(1, 4),
(1, 5),
(1, 6),
-- дополнительные её друзья
(1, 7),
(1, 8);

-- Bob's friends
INSERT INTO user_friends (user_id, friend_id) VALUES
-- общие с Alice
(2, 3),
(2, 4),
(2, 5),
(2, 6),
-- дополнительные его друзья
(2, 9),
(2, 10);

-- Сделаем также дружбы для других пользователей, чтобы было больше данных
INSERT INTO user_friends (user_id, friend_id) VALUES
-- Carol
(3, 1),
(3, 2),
(3, 4),
(3, 5),
-- Dave
(4, 1),
(4, 2),
(4, 3),
(4, 6),
-- Eve
(5, 1),
(5, 2),
(5, 6),
(5, 7),
-- Frank
(6, 1),
(6, 2),
(6, 4),
(6, 5),
-- Grace
(7, 1),
(7, 5),
(7, 8),
-- Heidi
(8, 1),
(8, 7),
(8, 9),
-- Ivan
(9, 2),
(9, 8),
(9, 10),
-- Judy
(10, 2),
(10, 9),
(10, 11),
-- и немного связей дальше
(11, 10),
(11, 12),
(12, 11),
(12, 13),
(13, 12),
(13, 14),
(14, 13),
(14, 15),
(15, 14),
(15, 16),
(16, 15),
(16, 17),
(17, 16),
(17, 18),
(18, 17),
(18, 19),
(19, 18),
(19, 20),
(20, 19);

