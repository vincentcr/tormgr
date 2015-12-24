
INSERT INTO users(id, email,password) VALUES
  ('86eb1856a155497aac7fd7ef50e7d2df', 'vincentcr@gmail.com', crypt('abcdefg', gen_salt('bf', 8)))
;

INSERT INTO access_tokens(secret, user_id, access) VALUES
  ('soverysecret', (SELECT id FROM users WHERE email = 'vincentcr@gmail.com'), 3)
;

INSERT INTO folders(id, owner_id, name) VALUES
  ('2307ebf7548c4cb7918f680787bf4760', '86eb1856a155497aac7fd7ef50e7d2df', 'films')
;
INSERT INTO torrents(folder_id, owner_id, url, info_hash) VALUES
  ('2307ebf7548c4cb7918f680787bf4760', '86eb1856a155497aac7fd7ef50e7d2df', 'https://torcache.net/torrent/0EFF4A129D185F91E28AE61CFAA2A4D2F5C6E429.torrent?title=[kat.cr]the.martian.2015.1080p.web.dl.dd5.1.h264.rarbg', '0EFF4A129D185F91E28AE61CFAA2A4D2F5C6E429'),
  ('2307ebf7548c4cb7918f680787bf4760', '86eb1856a155497aac7fd7ef50e7d2df', 'https://torcache.net/torrent/4A83E8481F5AB494B6936F0AF09AD02F116CD829.torrent?title=[kat.cr]steve.jobs.2015.dvdscr.xvid.ac3.hq.hive.cm8', '4A83E8481F5AB494B6936F0AF09AD02F116CD829'),
  ('2307ebf7548c4cb7918f680787bf4760', '86eb1856a155497aac7fd7ef50e7d2df', 'magnet:?xt=urn:btih:60B101018A32FBDDC264C1A2EB7B7E9A99DBFB6A&dn=mad+max+fury+road+2015+720p+brrip+x264+yify&tr=udp%3A%2F%2Ftracker.publicbt.com%2Fannounce&tr=udp%3A%2F%2Fglotorrents.pw%3A6969%2Fannounce', '60B101018A32FBDDC264C1A2EB7B7E9A99DBFB6A')
;

INSERT INTO folders(id, owner_id, name) VALUES
  ('3ed792dae7854f57ba455f33b1eeb371', '86eb1856a155497aac7fd7ef50e7d2df', 'tv')
;

INSERT INTO torrents(folder_id, owner_id, url, info_hash) VALUES
  ('3ed792dae7854f57ba455f33b1eeb371', '86eb1856a155497aac7fd7ef50e7d2df', 'https://torcache.net/torrent/59A4DDA27A8EC0BE2F91598C348A2405FD0F958D.torrent?title=[kat.cr]homeland.s05e12.web.dl.x264.fum.ettv', '59A4DDA27A8EC0BE2F91598C348A2405FD0F958D'),
  ('3ed792dae7854f57ba455f33b1eeb371', '86eb1856a155497aac7fd7ef50e7d2df', 'https://torcache.net/torrent/1FFA489D99F9C0761829D7A78F46DC272514240A.torrent?title=[kat.cr]the.affair.s02e12.hdtv.x264.killers.ettv', '1FFA489D99F9C0761829D7A78F46DC272514240A'),
  ('3ed792dae7854f57ba455f33b1eeb371', '86eb1856a155497aac7fd7ef50e7d2df', 'https://torcache.net/torrent/C2A67518FB965C88D821FBB0D4135068D580C3F7.torrent?title=[kat.cr]the.100.s01e01.720p.hdtv.x264.2hd.publichd', 'C2A67518FB965C88D821FBB0D4135068D580C3F7')
;

VACUUM ANALYZE;
