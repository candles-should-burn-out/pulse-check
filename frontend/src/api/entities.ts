export type Entity = {
  id: string;
  state: string;
};

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "/api";

export async function fetchEntities(signal?: AbortSignal): Promise<Entity[]> {
  const response = await fetch(`${API_BASE_URL}/entities`, {
    method: "GET",
    headers: {
      Accept: "application/json",
    },
    signal,
  });

  if (!response.ok) {
    throw new Error(`Не удалось загрузить сущности: HTTP ${response.status}`);
  }

  const payload: unknown = await response.json();

  if (!Array.isArray(payload)) {
    throw new Error("Сервер вернул неожиданный формат данных");
  }

  return payload.map((item) => {
    if (!isEntity(item)) {
      throw new Error("Сервер вернул сущность в неожиданном формате");
    }

    return item;
  });
}

function isEntity(value: unknown): value is Entity {
  return (
    typeof value === "object" &&
    value !== null &&
    "id" in value &&
    "state" in value &&
    typeof value.id === "string" &&
    typeof value.state === "string"
  );
}
