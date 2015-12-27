
--CREATE DOMAIN uid NOT NULL CHECK(VALUE ~ '^[a-fA-F0-9]{32}$');

CREATE OR REPLACE FUNCTION update_modified_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.date_modified = now();
    RETURN NEW;
END;
$$ language 'plpgsql';


CREATE TABLE users(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  email TEXT NOT NULL UNIQUE CHECK(email ~ '^[a-zA-Z0-9_%+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9][a-zA-Z0-9]+$'),
  password TEXT NOT NULL
);

CREATE UNIQUE INDEX idx_users_email ON users(lower(email));

CREATE TABLE access_tokens(
  secret TEXT PRIMARY KEY,
  user_id uuid REFERENCES users(id) NOT NULL,
  expires TIMESTAMP,
  access INT NOT NULL
);
CREATE INDEX idx_access_tokens_user_id ON access_tokens(user_id);
CREATE INDEX idx_access_tokens_expires ON access_tokens(expires) WHERE expires IS NOT NULL; -- index to use for deleting expired tokenss

CREATE TABLE folders(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  owner_id uuid REFERENCES users(id) NOT NULL,
  date_created TIMESTAMP NOT NULL DEFAULT timeofday()::TIMESTAMP,
  date_modified TIMESTAMP NOT NULL DEFAULT timeofday()::TIMESTAMP,
  name CITEXT NOT NULL CHECK (name != ''),
  UNIQUE(owner_id, name)
);
CREATE INDEX idx_folders_owner_id ON folders(owner_id);
CREATE INDEX idx_folders_id_owner_id ON folders(id, owner_id);
CREATE TRIGGER update_modtime BEFORE UPDATE ON folders FOR EACH ROW EXECUTE PROCEDURE update_modified_column();

CREATE TABLE torrents(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  folder CITEXT NOT NULL,
  owner_id uuid REFERENCES users(id) NOT NULL,
  date_created TIMESTAMP NOT NULL DEFAULT timeofday()::TIMESTAMP,
  date_modified TIMESTAMP NOT NULL DEFAULT timeofday()::TIMESTAMP,
  info_hash TEXT NOT NULL CHECK (info_hash ~ '^[A-F0-9]{40}$'),
  source_url TEXT,
  data BYTEA,
  status TEXT NOT NULL DEFAULT 'new' CHECK (status IN ('new', 'downloading', 'downloaded', 'failed')),
  FOREIGN KEY (folder, owner_id) REFERENCES folders(name, owner_id) ON DELETE RESTRICT ON UPDATE CASCADE,
  UNIQUE(id, info_hash)
);
CREATE INDEX idx_torrents_folder ON torrents(folder);
CREATE INDEX idx_torrents_id_owner_id ON torrents(id, owner_id);
CREATE TRIGGER update_modtime BEFORE UPDATE ON torrents FOR EACH ROW EXECUTE PROCEDURE update_modified_column();
