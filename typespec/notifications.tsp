import "@typespec/http"; 
import "@typespec/openapi3";

using TypeSpec.Http;

@doc("Базовая структура уведомления в системе NATS")
model Notification<T> {
  @doc("Уникальный идентификатор задачи")
  taskId: string;

  @doc("Тип события для маршрутизации на фронтенде")
  type: string;

  @doc("Версия схемы")
  version: "v1";

  @doc("Время создания события (ISO8601)")
  timestamp: utcDateTime;

  @doc("Полезная нагрузка события")
  payload: T;
}

@doc("Данные о прогрессе выполнения")
model ProgressPayload {
  @doc("Процент выполнения (0-100)")
  @minValue(0)
  @maxValue(100)
  progress: int32;

  @doc("Текстовое описание текущего этапа")
  stepName?: string;
}

@doc("Данные об ошибке")
model ErrorPayload {
  code: string;
  message: string;
}

// Конкретные модели сообщений
model ProgressNotification is Notification<ProgressPayload>;
model ErrorNotification is Notification<ErrorPayload>;

