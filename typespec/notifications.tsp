import "@typespec/http";
import "@typespec/openapi3";

using TypeSpec.Http;

@service(#{
  title: "Notification Service",
})

@doc("Сервис уведомлений")

namespace NotificationService {

  @doc("Базовая структура уведомления")
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
    @minValue(0)
    @maxValue(100)
    progress: int32;
    stepName?: string;
  }

  @doc("Данные об ошибке")
  model ErrorPayload {
    code: string;
    message: string;
  }

  model ProgressNotification is Notification<ProgressPayload>;
  model ErrorNotification is Notification<ErrorPayload>;

  @route("/sse-notifications")
  interface Notifications {
    @doc("Подключение к потоку SSE уведомлений для конкретного пользователя")
    @get
    subscribe(
      @doc("Идентификатор пользователя (например, rootcom)")
      @query user_id: string
    ): {
      @header contentType: "text/event-stream";
      @statusCode statusCode: 200;
      @body body: ProgressNotification | ErrorNotification;
    };
  }
}

