# Бенчмарки
Чисто для визуальной оценки, как парсер работает на разных платформах.\
Имеет смыслл смотреть только до строчки, `Database commit is ready` тк дальше идет disk IO, который интереса не представляет.
```
wget http://prog4food.altervista.org/sfiles/sample_it999.test
rm -rf ./bench/ chcache.db epgcache.tmp
zcat sample_it999.test | ./ott-play-epg-converter  -e "-,bench"
```

### Телевизор LG 2018 года
Результаты с mfpu=neon-vfpv4 одинаковые
```
2022-04-17T20:26:34 INF EPG compiler for OTT-play FOSS v0.3.3
2022-04-17T20:26:34 INF   git@prog4food (c) 2o22
2022-04-17T20:26:34 INF [bench] Read EPG from: StdIn
2022-04-17T20:26:34 INF [bench] EpgDb wiped 0.019636
2022-04-17T20:31:49 INF [bench] Epg parsing is ready 315.318471
2022-04-17T20:31:50 INF [bench] Database commit is ready 315.512321
2022-04-17T20:31:54 INF [bench] files count: 1
2022-04-17T20:31:54 ERR  error="database or disk is full"
2022-04-17T20:31:54 INF [bench] Json files is ready 319.556826
2022-04-17T20:31:54 INF Total Execution time: 319.561552
```

### A5x plus mini,  RK3328 (Armbian)
```
2022-04-17T16:45:08 INF EPG compiler for OTT-play FOSS v0.3.3
2022-04-17T16:45:08 INF   git@prog4food (c) 2o22
2022-04-17T16:45:08 INF Creating channel cache database...
2022-04-17T16:45:08 INF [bench] Read EPG from: StdIn
2022-04-17T16:45:08 INF [bench] EpgDb wiped 0.024552
2022-04-17T16:49:36 INF [bench] Epg parsing is ready 267.768850
2022-04-17T16:49:36 INF [bench] Database commit is ready 268.270171
2022-04-17T16:50:16 INF [bench] files count: 1678
2022-04-17T16:50:16 INF [bench] Json files is ready 308.475999
2022-04-17T16:50:16 INF Total Execution time: 308.506453
```

### PC, Ryzen 7 Pro 3700 (Windows 11)
```
2022-04-17T16:41:51 INF EPG compiler for OTT-play FOSS v0.3.3
2022-04-17T16:41:51 INF   git@prog4food (c) 2o22
2022-04-17T16:41:51 INF Creating channel cache database...
2022-04-17T16:41:51 INF [bench] Read EPG from: StdIn
2022-04-17T16:41:51 INF [bench] EpgDb wiped 0.112585
2022-04-17T16:42:15 INF [bench] Epg parsing is ready 23.668930
2022-04-17T16:42:18 INF [bench] Database commit is ready 26.857862
2022-04-17T16:42:24 INF [bench] files count: 1678
2022-04-17T16:42:24 INF [bench] Json files is ready 32.362656
2022-04-17T16:42:24 INF Total Execution time: 32.497998
```

### DIY NAS, Intel Atom 330 (X86_64 XPEnology)
```
2022-04-17T16:47:17 INF EPG compiler for OTT-play FOSS v0.3.3
2022-04-17T16:47:17 INF   git@prog4food (c) 2o22
2022-04-17T16:47:17 INF Creating channel cache database...
2022-04-17T16:47:17 INF [bench] Read EPG from: StdIn
2022-04-17T16:47:17 INF [bench] EpgDb wiped 0.022220
2022-04-17T16:52:19 INF [bench] Epg parsing is ready 301.611012
2022-04-17T16:52:22 INF [bench] Database commit is ready 304.469652
2022-04-17T16:53:15 INF [bench] files count: 1678
2022-04-17T16:53:15 INF [bench] Json files is ready 357.642238
2022-04-17T16:53:15 INF Total Execution time: 357.717273
```

### OnePlus 6, Termux (SDM845)
```
2022-04-17T15:36:43 INF EPG compiler for OTT-play FOSS v0.3.3
2022-04-17T15:36:43 INF   git@prog4food (c) 2o22
2022-04-17T15:36:43 INF [bench] Read EPG from: StdIn
2022-04-17T15:36:44 INF [bench] EpgDb wiped 0.209152
2022-04-17T15:37:42 INF [bench] Epg parsing is ready 58.208677
2022-04-17T15:37:42 INF [bench] Database commit is ready 58.350487
2022-04-17T15:37:51 INF [bench] files count: 1678
2022-04-17T15:37:51 INF [bench] Json files is ready 67.553936
2022-04-17T15:37:51 INF Total Execution time: 67.556929
```

### OnePlus 6, SDM845 (adb)
```
2022-04-17T15:38:40 INF EPG compiler for OTT-play FOSS v0.3.3
2022-04-17T15:38:40 INF   git@prog4food (c) 2o22
2022-04-17T15:38:40 INF Creating channel cache database...
2022-04-17T15:38:40 INF [bench] Read EPG from: StdIn
2022-04-17T15:38:40 INF [bench] EpgDb wiped 0.022165
2022-04-17T15:40:37 INF [bench] Epg parsing is ready 117.225780
2022-04-17T15:40:38 INF [bench] Database commit is ready 117.277786
2022-04-17T15:40:57 INF [bench] files count: 1678
2022-04-17T15:40:57 INF [bench] Json files is ready 136.599523
2022-04-17T15:40:57 INF Total Execution time: 136.628131
```

### Mini M8S Pro, Amlogic 912 (SlimBox)
```
2022-04-17T16:01:41 INF EPG compiler for OTT-play FOSS v0.3.3
2022-04-17T16:01:41 INF   git@prog4food (c) 2o22
2022-04-17T16:01:41 INF [bench] Read EPG from: StdIn
2022-04-17T16:01:41 INF [bench] EpgDb wiped 0.101458
2022-04-17T16:07:36 INF [bench] Epg parsing is ready 354.858614
2022-04-17T16:07:36 INF [bench] Database commit is ready 355.239313
2022-04-17T16:08:42 INF [bench] files count: 1678
2022-04-17T16:08:42 INF [bench] Json files is ready 420.984886
2022-04-17T16:08:42 INF Total Execution time: 420.988927
```

### BananaPi, AllWinner A20 (Armbian)
```
2022-04-17T20:45:56 INF EPG compiler for OTT-play FOSS v0.3.3
2022-04-17T20:45:56 INF   git@prog4food (c) 2o22
2022-04-17T20:45:56 INF Creating channel cache database...
2022-04-17T20:45:56 INF [bench] Read EPG from: StdIn
2022-04-17T20:45:56 INF [bench] EpgDb wiped 0.043520
2022-04-17T21:02:24 INF [bench] Epg parsing is ready 988.034203
2022-04-17T21:02:26 INF [bench] Database commit is ready 990.162280
2022-04-17T21:05:07 INF [bench] files count: 1678
2022-04-17T21:05:07 INF [bench] Json files is ready 1151.398450
2022-04-17T21:05:07 INF Total Execution time: 1151.500991
```