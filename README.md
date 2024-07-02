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
