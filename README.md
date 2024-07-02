# YouThumb

gRPC сервис для загрузки thumbnail изображений с YouTube.

## gRPC API

Пользователь отправляет запрос на получение thumbnail изображения по URL видео на YouTube. Сервис
возвращает изображение в виде последовательности чанков.

```proto
package youthumb.v1;

service ThumbnailService {
  rpc GetThumbnail(GetThumbnailRequest) returns (stream ThumbnailChunk);
}

message GetThumbnailRequest {
  string video_url = 1;
}

message ThumbnailChunk {
  string content_type = 1;
  bytes data = 2;
}
```

В качестве `video_url` можно использовать и обычную ссылку, и коротую. Например,
ссылки `https://www.youtube.com/watch?v=dQw4w9WgXcQ` и
`https://youtu.be/dQw4w9WgXcQ` эквивалентны. Больше поддерживаемых форматов можно
увидеть в тесте [`internal/thumbnail/url_test.go`](internal/thumbnail/url_test.go).

Полное определение сервиса и документация в файле [`proto/youthumb/v1/youthumb.proto`](proto/youthumb/v1/youthumb.proto).

## Архитектура

Две точки входа:

- [`cmd/server`](cmd/server) - точка входа для запуска gRPC сервера.
- [`cmd/client`](cmd/client) - пример gRPC клиента для отправки запросов на сервер.

Основной пакет с бизнес-логикой:

- [`internal/thumbnail`](internal/thumbnail) - пакет с бизнес-логикой сервиса и
  реализацией gRPC сервера.

Вспомогательныe пакеты для gRPC:

- [`internal/rpc`](internal/rpc) - пакет с основным конструктором gRPC сервера.
- [`internal/rpc/interceptor`](internal/rpc/interceptor) - пакет с middleware для gRPC сервера.
- [`internal/rpc/message`](internal/rpc/message) - пакет с общими сообщениями для gRPC сервера.

Вспомогательныe пакеты для приложения:

- [`internal/app`](internal/app) - пакет с общими компонентами приложения.
- [`internal/app/config`](internal/app/config) - пакет с конфигурацией приложения.
- [`internal/app/log`](internal/app/log) - пакет с логгером приложения.