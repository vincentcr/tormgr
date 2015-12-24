

CREATE OR REPLACE FUNCTION access_token_is_valid(TEXT, access_tokens) RETURNS BOOLEAN AS $$
  SELECT $2.secret = $1 AND ($2.expires IS NULL OR $2.expires > NOW())
$$ IMMUTABLE LANGUAGE SQL
;

--
--
-- CREATE OR REPLACE VIEW feed_json AS SELECT
--   id, owner_id, date_created, row_to_json(feed_json) as json
--   FROM (
--     SELECT REPLACE(id::text, '-', '') as id, owner_id,date_created, title, link,
--       (
--         SELECT COALESCE(array_to_json(array_agg(row_to_json(d))), '[]')
--         FROM (
--           SELECT REPLACE(id::text, '-', '') as id, link, title, description, date_added, date_modified
--           FROM feed_items
--           WHERE feed_id=feeds.id
--           ORDER BY date_added ASC
--         ) d
--       ) as items
--   FROM feeds
--   ) feed_json
-- ;
--
-- CREATE OR REPLACE VIEW feeds_json AS
--   SELECT json_agg(json) as json, owner_id
--   FROM (
--     SELECT * FROM feed_json ORDER BY date_created
--   ) AS feeds GROUP BY owner_id
-- ;
--
-- CREATE OR REPLACE FUNCTION feed_xml(uuid, uuid) RETURNS xml AS $$
--   SELECT
--     xmlelement(name "rss",
--       xmlattributes('2.0' as "version"),
--       xmlelement(name "channel",
--         xmlelement(name "link", feeds.link),
--         xmlelement(name "title", feeds.title),
--         xmlelement(name "description", coalesce(nullif(feeds.description, ''), feeds.title)),
--         (SELECT xmlagg(xmlelement(
--               name item,
--               xmlelement(name "link", feed_items.link),
--               xmlelement(name "title", feed_items.title),
--               xmlelement(name "guid", feeds.link || '/items/' || feed_items.id ),
--               xmlelement(name "pubDate", (SELECT to_char(feed_items.date_added, 'Dy, DD Mon YYYY HH24:MI:SS ') || 'GMT'))
--             ))
--             FROM (
--               SELECT * FROM feed_items
--                 WHERE feed_id=feeds.id
--                 ORDER BY date_added ASC
--             ) as feed_items
--         )
--       )
--     ) as xml
--   FROM feeds WHERE feeds.id=$1 AND feeds.owner_id=$2
-- $$ LANGUAGE SQL;
--
