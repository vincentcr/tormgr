

CREATE OR REPLACE FUNCTION access_token_is_valid(TEXT, access_tokens) RETURNS BOOLEAN AS $$
  SELECT $2.secret = $1 AND ($2.expires IS NULL OR $2.expires > NOW())
$$ IMMUTABLE LANGUAGE SQL
;
