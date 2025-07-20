CREATE TABLE IF NOT EXISTS ProgressDetail(
                    step_count int not null,
                    log text,
                    err text

);

CREATE TABLE IF NOT EXISTS NodeMinorUpgradeHistory(
                                    name text not null,
                                    last_updated timestamptz not null default now(),
                                    updated_version text not null unique,
--                                     progress_detail []text not null,
    updated_at timestamptz not null default now(),

    PRIMARY KEY (name)
    );

CREATE TABLE IF NOT EXISTS ClusterMinorUpgradeHistory(
    );

--     FOREIGN KEY (author_id) REFERENCES users(id)

CREATE INDEX IF NOT EXISTS idx_users_username ON users (username);
-- CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);

CREATE INDEX IF NOT EXISTS idx_articles_user_id ON articles (author_id);
CREATE INDEX IF NOT EXISTS idx_articles_slug ON articles (slug);

-- we can ADD indexes later IF we need them