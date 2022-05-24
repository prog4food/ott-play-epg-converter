# Бенчмарки
Чисто для визуальной оценки, как парсер работает на разных платформах.\
Имеет смыслл смотреть только до строчки, `Database commit is ready` тк дальше идет disk IO, который интереса не представляет.
```
wget http://prog4food.altervista.org/sfiles/sample_it999.test
rm -rf ./bench/ chcache.db epgcache.tmp
cat > bench.json << EOF
[{"id":"bench", "file":"sample_it999.test", "urls": ["http://prog4food.altervista.org/sfiles/sample_it999.test"]}]
EOF
./ott-play-epg-converter -c bench.json
```

### Телевизор LG 2018 года
```
2022-05-25T00:53:08 INF EPG converter for OTT-play FOSS dev-47740db
2022-05-25T00:53:08 INF   git@prog4food (c) 2o22
2022-05-25T00:53:08 INF chdb: create...
2022-05-25T00:53:09 INF [bench] Read EPG from: sample_it999.test
2022-05-25T00:53:09 INF [bench] input is gzipped
2022-05-25T00:53:09 INF [bench] Input ready 0.038794
2022-05-25T00:53:09 INF [bench] EpgDb ready 0.079022
2022-05-25T00:58:19 INF [bench] Epg parsing is ready 310.719587
2022-05-25T00:58:19 INF [bench] Database commit is ready 310.746613
2022-05-25T00:58:19 INF [bench] sorting epg_data...
2022-05-25T00:58:24 INF [bench] generating json...
2022-05-25T00:58:59 INF [bench] files count: 1677
2022-05-25T00:58:59 WRN [bench] has epg from the past!
2022-05-25T00:58:59 INF [bench] creating channels list...
2022-05-25T00:59:00 INF [bench] list ready, channels: 1677, with epg: 1677
2022-05-25T00:59:00 INF [bench] provider is ready 351.225684
2022-05-25T00:59:00 INF Total Execution time: 351.244613
```

### A5x plus mini,  RK3328 (Armbian)
```
2022-05-25T00:31:19 INF EPG converter for OTT-play FOSS dev-47740db
2022-05-25T00:31:19 INF   git@prog4food (c) 2o22
2022-05-25T00:31:19 INF chdb: create...
2022-05-25T00:31:19 WRN Cannot load providers.json
2022-05-25T00:31:19 INF [bench] Read EPG from: sample_it999.test
2022-05-25T00:31:19 INF [bench] input is gzipped
2022-05-25T00:31:19 INF [bench] Input ready 0.006755
2022-05-25T00:31:19 INF [bench] EpgDb ready 0.032169
2022-05-25T00:35:23 INF [bench] Epg parsing is ready 244.000439
2022-05-25T00:35:23 INF [bench] Database commit is ready 244.042669
2022-05-25T00:35:23 INF [bench] sorting epg_data...
2022-05-25T00:35:27 INF [bench] generating json...
2022-05-25T00:35:56 INF [bench] files count: 1677
2022-05-25T00:35:56 WRN [bench] has epg from the past!
2022-05-25T00:35:56 INF [bench] creating channels list...
2022-05-25T00:35:56 INF [bench] list ready, channels: 1677, with epg: 1677
2022-05-25T00:35:56 INF [bench] provider is ready 276.930434
2022-05-25T00:35:56 INF Total Execution time: 276.968937
```
и тест с выводом в файл tar.gz (gzip level: 1)
```
2022-05-25T00:37:21 INF EPG converter for OTT-play FOSS dev-47740db
2022-05-25T00:37:21 INF   git@prog4food (c) 2o22
2022-05-25T00:37:21 INF chdb: create...
2022-05-25T00:37:21 INF [bench] Read EPG from: sample_it999.test
2022-05-25T00:37:21 INF [bench] input is gzipped
2022-05-25T00:37:21 INF [bench] Input ready 0.049972
2022-05-25T00:37:21 INF [bench] EpgDb ready 0.080177
2022-05-25T00:41:26 INF [bench] Epg parsing is ready 245.336121
2022-05-25T00:41:26 INF [bench] Database commit is ready 245.367742
2022-05-25T00:41:26 INF [bench] sorting epg_data...
2022-05-25T00:41:30 INF [bench] generating json...
2022-05-25T00:42:00 INF [bench] files count: 1677
2022-05-25T00:42:00 WRN [bench] has epg from the past!
2022-05-25T00:42:00 INF [bench] creating channels list...
2022-05-25T00:42:00 INF [bench] list ready, channels: 1677, with epg: 1677
2022-05-25T00:42:00 INF [bench] provider is ready 279.256868
2022-05-25T00:42:00 INF Total Execution time: 279.320978
```

### PC, Ryzen 7 Pro 3700 (Windows 11)
```
2022-05-25T00:14:50 INF EPG converter for OTT-play FOSS dev-47740db
2022-05-25T00:14:50 INF   git@prog4food (c) 2o22
2022-05-25T00:14:50 INF chdb: create...
2022-05-25T00:14:50 INF [bench] Read EPG from: sample_it999.test
2022-05-25T00:14:50 INF [bench] input is gzipped
2022-05-25T00:14:50 INF [bench] Input ready 0.001733
2022-05-25T00:14:50 INF [bench] EpgDb ready 0.131677
2022-05-25T00:15:09 INF [bench] Epg parsing is ready 19.105446
2022-05-25T00:15:09 INF [bench] Database commit is ready 19.207240
2022-05-25T00:15:09 INF [bench] sorting epg_data...
2022-05-25T00:15:10 INF [bench] generating json...
2022-05-25T00:15:14 INF [bench] files count: 1677
2022-05-25T00:15:14 WRN [bench] has epg from the past!
2022-05-25T00:15:14 INF [bench] creating channels list...
2022-05-25T00:15:15 INF [bench] list ready, channels: 1677, with epg: 1677
2022-05-25T00:15:15 INF [bench] provider is ready 25.033948
2022-05-25T00:15:15 INF Total Execution time: 25.203066
```

### DIY NAS, Intel Atom 330 (X86_64 XPEnology)
```
2022-05-25T00:21:26 INF EPG converter for OTT-play FOSS dev-47740db
2022-05-25T00:21:26 INF   git@prog4food (c) 2o22
2022-05-25T00:21:26 INF chdb: create...
2022-05-25T00:21:26 WRN Cannot load providers.json
2022-05-25T00:21:26 INF [bench] Read EPG from: sample_it999.test
2022-05-25T00:21:26 INF [bench] input is gzipped
2022-05-25T00:21:26 INF [bench] Input ready 0.039902
2022-05-25T00:21:26 INF [bench] EpgDb ready 0.065457
2022-05-25T00:25:41 INF [bench] Epg parsing is ready 255.049383
2022-05-25T00:25:41 INF [bench] Database commit is ready 255.093095
2022-05-25T00:25:41 INF [bench] sorting epg_data...
2022-05-25T00:25:45 INF [bench] generating json...
2022-05-25T00:26:17 INF [bench] files count: 1677
2022-05-25T00:26:17 WRN [bench] has epg from the past!
2022-05-25T00:26:17 INF [bench] creating channels list...
2022-05-25T00:26:17 INF [bench] list ready, channels: 1677, with epg: 1677
2022-05-25T00:26:17 INF [bench] provider is ready 291.668253
2022-05-25T00:26:17 INF Total Execution time: 291.758175
```
и тест с выводом в файл tar.gz (gzip level: 1)
```
2022-05-25T00:38:09 INF EPG converter for OTT-play FOSS dev-47740db
2022-05-25T00:38:09 INF   git@prog4food (c) 2o22
2022-05-25T00:38:09 INF chdb: create...
2022-05-25T00:38:09 INF [bench] Read EPG from: sample_it999.test
2022-05-25T00:38:09 INF [bench] input is gzipped
2022-05-25T00:38:09 INF [bench] Input ready 0.039737
2022-05-25T00:38:09 INF [bench] EpgDb ready 0.093101
2022-05-25T00:42:24 INF [bench] Epg parsing is ready 255.432903
2022-05-25T00:42:24 INF [bench] Database commit is ready 255.454425
2022-05-25T00:42:24 INF [bench] sorting epg_data...
2022-05-25T00:42:28 INF [bench] generating json...
2022-05-25T00:43:01 INF [bench] files count: 1677
2022-05-25T00:43:01 WRN [bench] has epg from the past!
2022-05-25T00:43:01 INF [bench] creating channels list...
2022-05-25T00:43:02 INF [bench] list ready, channels: 1677, with epg: 1677
2022-05-25T00:43:02 INF [bench] provider is ready 293.106307
2022-05-25T00:43:02 INF Total Execution time: 293.215325
```

### OnePlus 6, Termux (SDM845)
```
2022-05-24T21:04:01 INF EPG converter for OTT-play FOSS dev-47740db
2022-05-24T21:04:01 INF   git@prog4food (c) 2o22
2022-05-24T21:04:01 INF chdb: create...
2022-05-24T21:04:01 WRN Cannot load providers.json
2022-05-24T21:04:01 INF [bench] Read EPG from: sample_it999.test
2022-05-24T21:04:01 INF [bench] input is gzipped
2022-05-24T21:04:01 INF [bench] Input ready 0.011495
2022-05-24T21:04:01 INF [bench] EpgDb ready 0.029029
2022-05-24T21:04:48 INF [bench] Epg parsing is ready 46.881120
2022-05-24T21:04:48 INF [bench] Database commit is ready 46.920858
2022-05-24T21:04:48 INF [bench] sorting epg_data...
2022-05-24T21:04:49 INF [bench] generating json...
2022-05-24T21:04:55 INF [bench] files count: 1677
2022-05-24T21:04:55 WRN [bench] has epg from the past!
2022-05-24T21:04:55 INF [bench] creating channels list...
2022-05-24T21:04:55 INF [bench] list ready, channels: 1677, with epg: 1677
2022-05-24T21:04:55 INF [bench] provider is ready 54.080849
2022-05-24T21:04:55 INF Total Execution time: 54.107098
```

### OnePlus 6, SDM845 (adb)
```
2022-05-24T21:06:51 INF EPG converter for OTT-play FOSS dev-47740db
2022-05-24T21:06:51 INF   git@prog4food (c) 2o22
2022-05-24T21:06:51 INF chdb: create...
2022-05-24T21:06:51 INF [bench] Read EPG from: sample_it999.test
2022-05-24T21:06:51 INF [bench] input is gzipped
2022-05-24T21:06:51 INF [bench] Input ready 0.002631
2022-05-24T21:06:51 INF [bench] EpgDb ready 0.019260
2022-05-24T21:09:33 INF [bench] Epg parsing is ready 161.634733
2022-05-24T21:09:33 INF [bench] Database commit is ready 161.649042
2022-05-24T21:09:33 INF [bench] sorting epg_data...
2022-05-24T21:09:36 INF [bench] generating json...
2022-05-24T21:09:55 INF [bench] files count: 1677
2022-05-24T21:09:55 WRN [bench] has epg from the past!
2022-05-24T21:09:55 INF [bench] creating channels list...
2022-05-24T21:09:55 INF [bench] list ready, channels: 1677, with epg: 1677
2022-05-24T21:09:55 INF [bench] provider is ready 183.888135
2022-05-24T21:09:55 INF Total Execution time: 183.910220
```

### Mini M8S Pro, Amlogic 912 (SlimBox)
```

```

### BananaPi, AllWinner A20 (Armbian)
```

```