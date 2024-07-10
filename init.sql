CREATE TABLE feed (
    pk SERIAL PRIMARY KEY,
    feed_url VARCHAR(256) UNIQUE NOT NULL
);
CREATE TABLE article (
    pk SERIAL PRIMARY KEY,
    title TEXT,
    content TEXT,
    source_url VARCHAR(256) UNIQUE NOT NULL,
    published TIMESTAMP WITH TIME ZONE,
    recorded TIMESTAMP WITH TIME ZONE DEFAULT transaction_timestamp(),
    feed_pk INT REFERENCES feed NOT NULL
);
CREATE TABLE person (
    pk UUID PRIMARY KEY,
    viewed TIMESTAMP WITH TIME ZONE DEFAULT transaction_timestamp()
);
CREATE TABLE subscribe (
    person_pk UUID REFERENCES person,
    feed_pk INT REFERENCES feed,
    UNIQUE (person_pk, feed_pk)
);
INSERT INTO person (pk)
VALUES
('35be0a7c-8570-4987-be59-efeac5906d74'),
('093d7936-8c3e-4e36-aa4a-8fe2e820ea4f');

INSERT INTO feed (feed_url)
VALUES
('https://defensivecss.dev/feed/feed.xml'),
('https://ariya.io/index.xml'),
('https://justmarkup.com/feed/feed.xml'),
('https://www.andreaverlicchi.eu/feed/feed.xml'),
('https://blog.almaer.com/feed/'),
('https://blog.shhdharmen.me/rss.xml'),
('https://jwdallas.com/feed.xml'),
('https://www.konnorrogers.com/feed.xml'),
('https://daniel.do/rss.xml'),
('https://trentwalton.com/feed.xml'),
('https://www.coolfields.co.uk/feed/'),
('https://webplatform.news/feed.xml');



