export type StatusRole = "status_owner" | "assistant";

export type StatusDefinition = {
  id: string;
  name: string;
  border_color: string;
  background_color: string;
  text_color: string;
  created_at: string;
  updated_at: string;
};

export type StatusSet = {
  id: string;
  owner_user_id: string;
  role: StatusRole;
  statuses: StatusDefinition[];
};

export type StatusInput = {
  name: string;
  border_color: string;
  background_color: string;
  text_color: string;
};

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "/api";

export async function fetchStatusSet(
  accessToken: string,
  signal?: AbortSignal
): Promise<StatusSet> {
  const payload = await requestJSON("/status-set", accessToken, { signal });

  if (!isStatusSet(payload)) {
    throw new Error("Сервер вернул набор статусов в неожиданном формате");
  }

  return payload;
}

export async function createStatus(
  accessToken: string,
  input: StatusInput,
  signal?: AbortSignal
): Promise<StatusDefinition> {
  const payload = await requestJSON("/statuses", accessToken, {
    method: "POST",
    body: JSON.stringify(input),
    signal,
  });

  if (!isStatusDefinition(payload)) {
    throw new Error("Сервер вернул статус в неожиданном формате");
  }

  return payload;
}

export async function updateStatus(
  accessToken: string,
  statusId: string,
  input: StatusInput,
  signal?: AbortSignal
): Promise<StatusDefinition> {
  const payload = await requestJSON(`/statuses/${statusId}`, accessToken, {
    method: "PATCH",
    body: JSON.stringify(input),
    signal,
  });

  if (!isStatusDefinition(payload)) {
    throw new Error("Сервер вернул статус в неожиданном формате");
  }

  return payload;
}

export async function deleteStatus(
  accessToken: string,
  statusId: string,
  signal?: AbortSignal
) {
  await requestJSON(`/statuses/${statusId}`, accessToken, {
    method: "DELETE",
    signal,
  });
}

async function requestJSON(
  path: string,
  accessToken: string,
  init: RequestInit = {}
): Promise<unknown> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    headers: {
      Accept: "application/json",
      Authorization: `Bearer ${accessToken}`,
      ...(init.body ? { "Content-Type": "application/json" } : {}),
      ...init.headers,
    },
  });

  if (response.status === 204) {
    return null;
  }

  const payload: unknown = await response.json().catch(() => null);

  if (!response.ok) {
    const errorCode =
      typeof payload === "object" &&
      payload !== null &&
      "error" in payload &&
      typeof payload.error === "string"
        ? payload.error
        : `HTTP ${response.status}`;
    throw new Error(statusErrorMessage(errorCode));
  }

  return payload;
}

function statusErrorMessage(errorCode: string) {
  switch (errorCode) {
    case "status_limit_reached":
      return "Достигнут лимит статусов";
    case "status_set_read_only":
      return "Набор статусов доступен только для чтения";
    case "invalid_status":
      return "Проверьте имя и цвета статуса";
    case "status_not_found":
      return "Статус не найден";
    default:
      return `Не удалось выполнить запрос: ${errorCode}`;
  }
}

function isStatusSet(value: unknown): value is StatusSet {
  return (
    typeof value === "object" &&
    value !== null &&
    "id" in value &&
    "owner_user_id" in value &&
    "role" in value &&
    "statuses" in value &&
    typeof value.id === "string" &&
    typeof value.owner_user_id === "string" &&
    (value.role === "status_owner" || value.role === "assistant") &&
    Array.isArray(value.statuses) &&
    value.statuses.every(isStatusDefinition)
  );
}

function isStatusDefinition(value: unknown): value is StatusDefinition {
  return (
    typeof value === "object" &&
    value !== null &&
    "id" in value &&
    "name" in value &&
    "border_color" in value &&
    "background_color" in value &&
    "text_color" in value &&
    "created_at" in value &&
    "updated_at" in value &&
    typeof value.id === "string" &&
    typeof value.name === "string" &&
    typeof value.border_color === "string" &&
    typeof value.background_color === "string" &&
    typeof value.text_color === "string" &&
    typeof value.created_at === "string" &&
    typeof value.updated_at === "string"
  );
}
