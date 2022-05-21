# EPG converter for OTT-play FOSS

# Описание
Инструмент создания телепрограммы для OTT-Play FOSS, использует 1 поток, и буферное чтение из файла, что позволяет минимизировать нагрузку на систему при разборе больших файлов, потребляет примерно 25 мб оперативной памяти при обработке 300 МБ файла, на современном процессоре тратит на это порядка 35 сек. Для любопытных есть бенчмарки.

Вся информация о провайдерах заносится в конфиг файлы в формате json, можно настроить провайдер так, что он будет получать данные через stdin (а значит файл можно вообще не скачивать на диск). Пример конфига в `samples/sample_config.json`

Поддерживается прямое скачивание файлов с поддержкой редиректов(301/302), а также прямая работа с gzip источниками.

После каждого запуска актуализируется база данных по обработанным провайдерам `chcache.db` на ее основе в конце создается общий список каналов.\
`epgcache.tmp` это данные epg последнего обработанного провайдера, пересоздается при обработке нового провайдера.\
Файл `providers.json` содержит имя провайдера, хеши поддерживаемых url-tvg, и время последней передачи в EPG (для контроля актуальности провайдера).\
Для каждого провайдера также создается `channels.json`, в котором содержится хеш ид канала, время последней передачи на канале (позволяет контролировать актуальность и наличие EPG), ссылку на логотип канала, список названий канала.


# Аргументы командной строки
```
Общая схема: <app> [--epg-ram] [-c OPTS]
Основные опции:
    --epg-ram  включает обработку базы в оперативной памяти, дает ~20% прирост
               производительности, но увеличивает потребление памяти до ~200МБ
    -c <opts>  обработать XMLTV файл по параметрам из конфиг файла
  ПРИМЕЧАНИЕ: Символ "," используется как разделитель в блоке <opts>,
              аргументы -c можно комбинировать

  -c config_file[,prov_name]
    config_file  файл со списком для обработки
    prov_name    дополнительный фильтр, для выборки только одного провайдера
```


# Конфиг
Представляет из себя массив с json объектами, у которых могут быть следующие свойства:
 * `id` - идентификатор (короткое имя) провайдера, для внутреннего использования плеером;
 * `file` - имя файла на диске, который необходимо обрабатывать, в качестве имени
    файла можно указать `-`, тогда будет читаться ввод из StdIn, и можно загружать XML из cat/zcat/curl в различных комбинациях. *Логично, что не стоит пытаться обработать 2 источника с `-` за один запуск*. **Имеет приоритет над urls, при указании обоих, предпочтение будет отдаваться file**;
 * `urls` - массив ссылок, которые могут указываться в плейлистах (url-tvg), для этого провайдера, самая первая ссылка скачивается и обрабатывается (если не определено свойство *file*), остальные используются для определения и сопоставления провайдеров, схема ссылки (http/https) при сопоставлении не учитывается;

Обязательными являются только 2 свойства: `id`, `url` или `file`

### Примеры конфига
* Провайдер `it999f` с источником из файла `it999.xml.gz`:
  ```json
  {"id":"it999f", "file":"it999.xml.gz"}
  ```
* Провайдер `it999u` с источником из ссылки `https://epg.it999.ru/epg.xml.gz`:
  ```json
  {"id":"it999u", "urls": ["https://epg.it999.ru/epg.xml.gz", "http://epg.it999.ru/epg.xml.gz"]}
  ```
* Провайдер `lightiptv` с источником из ссылки `https://epg.lightiptv.cc/epg.xml.gz`, загрузка из файла не будет использована, тк свойство записано как `_file`:
  ```json
  {"id":"lightiptv", "_file":"lightiptv.xml.gz", "urls": ["https://epg.lightiptv.cc/epg.xml.gz", "http://epg.lightiptv.cc/epg.xml.gz"] }
  ```
* Провайдер `iptvx.one` с источником из StdIn, также будет автоматически определяться у клиентов ottg.tv:
  ```json
    {"id":"iptvx.one", "file":"-", "urls": ["https://iptvx.one/EPG", "https://ottg.tv/epg.xml.gz"]}
  ```


# Примеры запуска
```
### Создать EPG для провайдера it999 из конфиг файла sample_config.json:
  ott-play-epg-converter -c sample_config.json,it999
### Создать EPG для провайдера intest из конфиг файла sample_config.json:
  zcat somepg.xml.gz | ott-play-epg-converter -c sample_config.json,intest
### Создать EPG для всех провайдеров из конфиг файла sample_config.json:
  ott-play-epg-converter -c sample_config.json
### Другие примеры:
  cat epgone.xml | ott-play-epg-converter -c sample_config.json,intest
  curl --silent http://prov.host/epg.xml.gz | gzip -d -c - | ott-play-epg-converter -c sample_config.json,intest
  curl --silent --compressed http://prov.host/epg.xml | ott-play-epg-converter -c sample_config.json,intest
  ...
  ott-play-epg-converter -c sample_config.json,it999
```


# Prebuild
Готовые бинарники я компилирую для Windows (x64, x86)/Linux (x64, arm64, arm7a-soft-float, arm7-hard-float)\
Иногда буду добавлять версии для Android (arm64, arm7a-soft-float)\
В теории можно завести на всем, что поддерживает GO и кросс-компил, но сборкой придется заняться самостоятельно.


# Бенчмарки
В файле [benchmarks.md](benchmarks.md) есть немного тестов EPG от it999 на различных платформах.