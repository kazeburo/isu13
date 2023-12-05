
index

```
mysql> alter table livestream_tags add index idx_livestream_id (livestream_id);
mysql> alter table livestream_tags add index idx_tag_id_livestream_id (tag_id, livestream_id);
mysql> alter table livestreams add index idx_user_id (user_id);
mysql> alter table icons add index idx_user_id (user_id);
mysql> alter table livecomments add index idx_livestream_id (livestream_id);
mysql> alter table themes add index idx_user_id (user_id);
mysql> alter table reactions add index idx_livestream_id(livestream_id,created_at);
mysql> alter table livestream_viewers_history add index idx_user_id_livestream_id  (user_id, livestream_id);
mysql> alter table ng_words add index idx_livestream_id_user_id (livestream_id,user_id);
mysql> alter table reservation_slots add index idx_start_end (start_at, end_at);
mysql> alter table livecomment_reports add index idx_livestream_id(livestream_id);
```


livestream_tags

```
mysql> ALTER TABLE livestreams ADD raw_tags VARCHAR(255) NOT NULL DEFAULT "" AFTER end_at;
mysql> UPDATE livestreams u SET raw_tags=IFNULL((SELECT GROUP_CONCAT(tag_id SEPARATOR ",") FROM livestream_tags WHERE livestream_id=u.id GROUP BY livestream_id),"");
```

create icon_hash on users

```
ALTER TABLE users ADD icon_hash VARCHAR(255) NOT NULL DEFAULT "" AFTER description;
```


move theme to users

```
ALTER TABLE users ADD dark_mode BOOLEAN NOT NULL DEFAULT FALSE AFTER icon_hash;
UPDATE users u JOIN themes t ON u.id = t.user_id SET u.dark_mode = t.dark_mode;
```

score a5a8fc9d3d362873d420309521a47fe49fa67284

```
2023-11-30T02:26:18.335Z        info    staff-logger    bench/bench.go:301      シナリオカウンタを出力します
2023-11-30T02:26:18.335Z        info    isupipe-benchmarker     配信を最後まで視聴できた視聴者数        {"viewers": 787}
2023-11-30T02:26:18.335Z        info    staff-logger    bench/bench.go:323      [シナリオ aggressive-streamer-moderate] 8 回成功
2023-11-30T02:26:18.335Z        info    staff-logger    bench/bench.go:323      [シナリオ streamer-cold-reserve] 1146 回成功
2023-11-30T02:26:18.335Z        info    staff-logger    bench/bench.go:323      [シナリオ streamer-moderate] 561 回成功, 9 回失敗
2023-11-30T02:26:18.335Z        info    staff-logger    bench/bench.go:323      [シナリオ viewer-report] 58 回成功
2023-11-30T02:26:18.335Z        info    staff-logger    bench/bench.go:323      [シナリオ viewer-spam] 10 回成功
2023-11-30T02:26:18.335Z        info    staff-logger    bench/bench.go:323      [シナリオ viewer] 787 回成功
2023-11-30T02:26:18.335Z        info    staff-logger    bench/bench.go:323      [失敗シナリオ streamer-moderate-fail] 9 回失敗
2023-11-30T02:26:18.335Z        info    staff-logger    bench/bench.go:329      DNSAttacker並列数: 2
2023-11-30T02:26:18.335Z        info    staff-logger    bench/bench.go:330      名前解決成功数: 426
2023-11-30T02:26:18.335Z        info    staff-logger    bench/bench.go:331      名前解決失敗数: 115
2023-11-30T02:26:18.335Z        info    staff-logger    bench/bench.go:335      スコア: 153855
```

score 57779b138135e7c148755439a479fd632d07a982

```
2023-11-30T12:04:10.379Z        info    staff-logger    bench/bench.go:301      シナリオカウンタを出力します
2023-11-30T12:04:10.379Z        info    isupipe-benchmarker     配信を最後まで視聴できた視聴者数        {"viewers": 1119}
2023-11-30T12:04:10.379Z        info    staff-logger    bench/bench.go:323      [シナリオ aggressive-streamer-moderate] 18 回成功
2023-11-30T12:04:10.379Z        info    staff-logger    bench/bench.go:323      [シナリオ streamer-cold-reserve] 1261 回成功, 30 回失敗
2023-11-30T12:04:10.379Z        info    staff-logger    bench/bench.go:323      [シナリオ streamer-moderate] 574 回成功, 24 回失敗
2023-11-30T12:04:10.379Z        info    staff-logger    bench/bench.go:323      [シナリオ viewer-report] 58 回成功
2023-11-30T12:04:10.379Z        info    staff-logger    bench/bench.go:323      [シナリオ viewer-spam] 20 回成功
2023-11-30T12:04:10.379Z        info    staff-logger    bench/bench.go:323      [シナリオ viewer] 1119 回成功
2023-11-30T12:04:10.379Z        info    staff-logger    bench/bench.go:323      [失敗シナリオ streamer-cold-reserve-fail] 30 回失敗
2023-11-30T12:04:10.379Z        info    staff-logger    bench/bench.go:323      [失敗シナリオ streamer-moderate-fail] 24 回失敗
2023-11-30T12:04:10.379Z        info    staff-logger    bench/bench.go:329      DNSAttacker並列数: 2
2023-11-30T12:04:10.379Z        info    staff-logger    bench/bench.go:330      名前解決成功数: 445
2023-11-30T12:04:10.379Z        info    staff-logger    bench/bench.go:331      名前解決失敗数: 114
2023-11-30T12:04:10.379Z        info    staff-logger    bench/bench.go:335      スコア: 218103
```

score f13772f7e436b0b2dffb6c4f7cef84885af82fa6

```
2023-12-01T12:57:51.593Z        info    staff-logger    bench/bench.go:301      シナリオカウンタを出力します
2023-12-01T12:57:51.593Z        info    isupipe-benchmarker     配信を最後まで視聴できた視聴者数        {"viewers": 1402}
2023-12-01T12:57:51.593Z        info    staff-logger    bench/bench.go:323      [シナリオ aggressive-streamer-moderate] 15 回成功, 2 回失敗
2023-12-01T12:57:51.593Z        info    staff-logger    bench/bench.go:323      [シナリオ streamer-cold-reserve] 1272 回成功, 38 回失敗
2023-12-01T12:57:51.593Z        info    staff-logger    bench/bench.go:323      [シナリオ streamer-moderate] 696 回成功, 26 回失敗
2023-12-01T12:57:51.593Z        info    staff-logger    bench/bench.go:323      [シナリオ viewer-report] 58 回成功
2023-12-01T12:57:51.593Z        info    staff-logger    bench/bench.go:323      [シナリオ viewer-spam] 17 回成功
2023-12-01T12:57:51.593Z        info    staff-logger    bench/bench.go:323      [シナリオ viewer] 1402 回成功, 10 回失敗
2023-12-01T12:57:51.593Z        info    staff-logger    bench/bench.go:323      [失敗シナリオ aggressive-streamer-moderate-fail] 2 回失敗
2023-12-01T12:57:51.593Z        info    staff-logger    bench/bench.go:323      [失敗シナリオ streamer-cold-reserve-fail] 38 回失敗
2023-12-01T12:57:51.593Z        info    staff-logger    bench/bench.go:323      [失敗シナリオ streamer-moderate-fail] 26 回失敗
2023-12-01T12:57:51.593Z        info    staff-logger    bench/bench.go:323      [失敗シナリオ viewer-fail] 10 回失敗
2023-12-01T12:57:51.593Z        info    staff-logger    bench/bench.go:329      DNSAttacker並列数: 2
2023-12-01T12:57:51.593Z        info    staff-logger    bench/bench.go:330      名前解決成功数: 427
2023-12-01T12:57:51.593Z        info    staff-logger    bench/bench.go:331      名前解決失敗数: 112
2023-12-01T12:57:51.593Z        info    staff-logger    bench/bench.go:335      スコア: 272355
```

```
CREATE TABLE `livestream_score` (
  `livestream_id` BIGINT NOT NULL PRIMARY KEY,
  `user_id` BIGINT NOT NULL,
  `score` BIGINT NOT NULL DEFAULT '0'
) ENGINE=InnoDB CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;

alter table livestream_score add index idx_user_id(user_id,score);

DELIMITER $$
CREATE TRIGGER insert_livestreams
AFTER INSERT ON livestreams
FOR EACH ROW
BEGIN
INSERT INTO livestream_score (livestream_id, user_id, score) VALUES (NEW.id, NEW.user_id, 0);
END
$$

DELIMITER $$
CREATE TRIGGER insert_livecomments
AFTER INSERT ON livecomments
FOR EACH ROW
BEGIN
UPDATE livestream_score SET score = score + NEW.tip WHERE livestream_id = NEW.livestream_id;
END
$$

DELIMITER $$
CREATE TRIGGER delete_livecomments
AFTER DELETE ON livecomments
FOR EACH ROW
BEGIN
UPDATE livestream_score SET score = score - OLD.tip WHERE livestream_id = OLD.livestream_id;
END
$$

DELIMITER $$
CREATE TRIGGER insert_reactions
AFTER INSERT ON reactions
FOR EACH ROW
BEGIN
UPDATE livestream_score SET score = score + 1 WHERE livestream_id = NEW.livestream_id;
END
$$

DELIMITER $$
CREATE TRIGGER delete_reactions
AFTER DELETE ON reactions
FOR EACH ROW
BEGIN
UPDATE livestream_score SET score = score - 1 WHERE livestream_id = OLD.livestream_id;
END
$$

```

score 4fef946f98d3e9f81bb0e8ecdd88d2fedbedab73

```
2023-12-03T13:57:21.474Z        info    isupipe-benchmarker     配信を最後まで視聴できた視聴者数        {"viewers": 1454}
2023-12-03T13:57:21.474Z        info    staff-logger    bench/bench.go:323      [シナリオ aggressive-streamer-moderate] 10 回成功, 2 回失敗
2023-12-03T13:57:21.474Z        info    staff-logger    bench/bench.go:323      [シナリオ dns-watertorture-attack] 1 回成功
2023-12-03T13:57:21.474Z        info    staff-logger    bench/bench.go:323      [シナリオ streamer-cold-reserve] 1291 回成功, 53 回失敗
2023-12-03T13:57:21.474Z        info    staff-logger    bench/bench.go:323      [シナリオ streamer-moderate] 701 回成功, 17 回失敗
2023-12-03T13:57:21.474Z        info    staff-logger    bench/bench.go:323      [シナリオ viewer-report] 59 回成功
2023-12-03T13:57:21.474Z        info    staff-logger    bench/bench.go:323      [シナリオ viewer-spam] 12 回成功
2023-12-03T13:57:21.474Z        info    staff-logger    bench/bench.go:323      [シナリオ viewer] 1454 回成功, 10 回失敗
2023-12-03T13:57:21.474Z        info    staff-logger    bench/bench.go:323      [失敗シナリオ aggressive-streamer-moderate-fail] 2 回失敗
2023-12-03T13:57:21.474Z        info    staff-logger    bench/bench.go:323      [失敗シナリオ streamer-cold-reserve-fail] 53 回失敗
2023-12-03T13:57:21.474Z        info    staff-logger    bench/bench.go:323      [失敗シナリオ streamer-moderate-fail] 17 回失敗
2023-12-03T13:57:21.474Z        info    staff-logger    bench/bench.go:323      [失敗シナリオ viewer-fail] 10 回失敗
2023-12-03T13:57:21.474Z        info    staff-logger    bench/bench.go:329      DNSAttacker並列数: 2
2023-12-03T13:57:21.474Z        info    staff-logger    bench/bench.go:330      名前解決成功数: 424
2023-12-03T13:57:21.474Z        info    staff-logger    bench/bench.go:331      名前解決失敗数: 118
2023-12-03T13:57:21.474Z        info    staff-logger    bench/bench.go:335      スコア: 282495
```

score f898e7d2a1ae4a8392c18916a49954a7170c3d38

```
2023-12-04T11:54:56.871Z        info    staff-logger    bench/bench.go:301      シナリオカウンタを出力します
2023-12-04T11:54:56.871Z        info    isupipe-benchmarker     配信を最後まで視聴できた視聴者数        {"viewers": 1604}
2023-12-04T11:54:56.871Z        info    staff-logger    bench/bench.go:323      [シナリオ aggressive-streamer-moderate] 11 回成功
2023-12-04T11:54:56.871Z        info    staff-logger    bench/bench.go:323      [シナリオ streamer-cold-reserve] 1303 回成功, 60 回失敗
2023-12-04T11:54:56.871Z        info    staff-logger    bench/bench.go:323      [シナリオ streamer-moderate] 761 回成功, 20 回失敗
2023-12-04T11:54:56.871Z        info    staff-logger    bench/bench.go:323      [シナリオ viewer-report] 59 回成功
2023-12-04T11:54:56.871Z        info    staff-logger    bench/bench.go:323      [シナリオ viewer-spam] 13 回成功
2023-12-04T11:54:56.871Z        info    staff-logger    bench/bench.go:323      [シナリオ viewer] 1604 回成功
2023-12-04T11:54:56.871Z        info    staff-logger    bench/bench.go:323      [失敗シナリオ streamer-cold-reserve-fail] 60 回失敗
2023-12-04T11:54:56.871Z        info    staff-logger    bench/bench.go:323      [失敗シナリオ streamer-moderate-fail] 20 回失敗
2023-12-04T11:54:56.871Z        info    staff-logger    bench/bench.go:329      DNSAttacker並列数: 2
2023-12-04T11:54:56.871Z        info    staff-logger    bench/bench.go:330      名前解決成功数: 440
2023-12-04T11:54:56.871Z        info    staff-logger    bench/bench.go:331      名前解決失敗数: 117
2023-12-04T11:54:56.871Z        info    staff-logger    bench/bench.go:335      スコア: 311636
```

score HEAD

```
2023-12-05T15:52:59.559Z        info    staff-logger    bench/bench.go:301      シナリオカウンタを出力します
2023-12-05T15:52:59.559Z        info    isupipe-benchmarker     配信を最後まで視聴できた視聴者数        {"viewers": 2725}
2023-12-05T15:52:59.559Z        info    staff-logger    bench/bench.go:323      [シナリオ aggressive-streamer-moderate] 14 回成功, 2 回失敗
2023-12-05T15:52:59.559Z        info    staff-logger    bench/bench.go:323      [シナリオ streamer-cold-reserve] 628 回成功, 7234 回失敗
2023-12-05T15:52:59.559Z        info    staff-logger    bench/bench.go:323      [シナリオ streamer-moderate] 1540 回成功, 58 回失敗
2023-12-05T15:52:59.559Z        info    staff-logger    bench/bench.go:323      [シナリオ viewer-report] 59 回成功
2023-12-05T15:52:59.559Z        info    staff-logger    bench/bench.go:323      [シナリオ viewer-spam] 16 回成功
2023-12-05T15:52:59.559Z        info    staff-logger    bench/bench.go:323      [シナリオ viewer] 2725 回成功, 10 回失敗
2023-12-05T15:52:59.559Z        info    staff-logger    bench/bench.go:323      [失敗シナリオ aggressive-streamer-moderate-fail] 2 回失敗
2023-12-05T15:52:59.559Z        info    staff-logger    bench/bench.go:323      [失敗シナリオ streamer-cold-reserve-fail] 7234 回失敗
2023-12-05T15:52:59.559Z        info    staff-logger    bench/bench.go:323      [失敗シナリオ streamer-moderate-fail] 58 回失敗
2023-12-05T15:52:59.559Z        info    staff-logger    bench/bench.go:323      [失敗シナリオ viewer-fail] 10 回失敗
2023-12-05T15:52:59.559Z        info    staff-logger    bench/bench.go:329      DNSAttacker並列数: 2
2023-12-05T15:52:59.559Z        info    staff-logger    bench/bench.go:330      名前解決成功数: 497
2023-12-05T15:52:59.559Z        info    staff-logger    bench/bench.go:331      名前解決失敗数: 118
2023-12-05T15:52:59.559Z        info    staff-logger    bench/bench.go:335      スコア: 390127
```

