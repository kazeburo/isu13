
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

score HEAD

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
