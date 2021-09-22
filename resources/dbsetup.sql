DROP TABLE IF EXISTS gatomemes;
CREATE TABLE gatomemes (
  id        INT UNSIGNED AUTO_INCREMENT,
  line1     VARCHAR(255) NOT NULL,
  line2     VARCHAR(255) NOT NULL,
  PRIMARY KEY (id),
  CONSTRAINT unique_lines UNIQUE (line1, line2)
);

INSERT INTO gatomemes 
  (line1, line2)
VALUES 
  ('LINE1 TEXT1', 'LINE2 TEXT1'),
  ('LINE1 TEXT2', 'LINE2 TEXT2'),
  ('LINE1 TEXT3', 'LINE2 TEXT3');

DROP TABLE IF EXISTS image_data;
CREATE TABLE image_data (
  id          INT UNSIGNED AUTO_INCREMENT NOT NULL,
  text_id     INT UNSIGNED NOT NULL,
  file_name   VARCHAR(255),
  likes       INT UNSIGNED NOT NULL DEFAULT 0,
  dislikes    INT UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (id)
);

DROP TABLE IF EXISTS comments;
CREATE TABLE comments (
  id          INT UNSIGNED AUTO_INCREMENT NOT NULL,
  image_id    INT UNSIGNED NOT NULL,
  identity    VARCHAR(36),
  time        TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  comment     VARCHAR(255),
  PRIMARY KEY (id)
);

INSERT INTO image_data (text_id)
SELECT id FROM gatomemes ORDER BY id;

DROP TABLE IF EXISTS user;
CREATE TABLE user (
  identity        VARCHAR(36) NOT NULL,
  user_name       VARCHAR(20) NOT NULL,
  password        VARCHAR(255) NOT NULL,
  session_key     VARCHAR(36),
  is_disabled     BOOL NOT NULL DEFAULT 0,
  is_admin        BOOL NOT NULL DEFAULT 0,
  is_root         BOOL NOT NULL DEFAULT 0,
  reg_time        TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  last_seen_time  TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY(identity),
  UNIQUE(user_name),
  UNIQUE(session_key)
);