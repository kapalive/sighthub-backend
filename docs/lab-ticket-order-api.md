# Lab Ticket Order API

## Unified Order — один endpoint, бэкенд роутит сам

Бэкенд определяет провайдера по `lab_id` тикета и `source` линзы в invoice.

| Source линзы | Провайдер | Электронный заказ |
|---|---|---|
| `vision_web` | VisionWeb SOAP | Да |
| `zeiss_only` | Zeiss REST API | Да |
| `custom` | — | Нет (ручной) |

---

## 1. Проверка готовности

```
GET /api/ticket/{ticket_id}/order-requirements
```

Ответ всегда содержит:

| Поле | Тип | Описание |
|---|---|---|
| `lens_source` | string | `"vision_web"` / `"zeiss_only"` / `"custom"` / `""` |
| `can_order` | bool | Можно ли отправить электронно |
| `provider` | string | `"vision_web"` / `"zeiss"` (только если can_order=true) |
| `ready` | bool | Все обязательные поля заполнены |
| `fields` | array | Список полей с filled/required |

### Custom lens:
```json
{
  "ready": false,
  "lens_source": "custom",
  "can_order": false,
  "message": "Manual order — electronic submission not available for custom lenses"
}
```

### VisionWeb:
```json
{
  "ready": true,
  "lens_source": "vision_web",
  "can_order": true,
  "provider": "vision_web",
  "fields": [
    { "field": "lab_id", "label": "Laboratory", "required": true, "filled": true, "value": "TRI SUPREME" },
    { "field": "vw_account", "label": "VisionWeb Account", "required": true, "filled": true, "value": "slo=1161 bill=12870 ship=12870" },
    { "field": "vw_design_code", "label": "Lens Design", "required": true, "filled": true, "value": "COMFORT2" },
    { "field": "vw_material_code", "label": "Lens Material", "required": true, "filled": true, "value": "PL-50-NONE-NONE-16" },
    { "field": "od_sph", ... },
    ...
  ]
}
```

### Zeiss:
```json
{
  "ready": false,
  "lens_source": "zeiss_only",
  "can_order": true,
  "provider": "zeiss",
  "fields": [
    { "field": "zeiss_auth", "label": "Zeiss Authentication", "required": true, "filled": false },
    { "field": "customer_number", "label": "Zeiss Customer Number", "required": true, "filled": false },
    { "field": "lab_id", "label": "Laboratory (CARL ZEISS)", "required": true, "filled": true, "value": "CARL ZEISS" },
    { "field": "commercial_code", "label": "Lens Commercial Code", "required": true, "filled": true, "value": "47391" },
    { "field": "coating_code", "label": "Coating", "required": false, "filled": true, "value": "DD" },
    { "field": "od_sph", "label": "OD Sphere", "required": true, "filled": true, "value": "-1.00" },
    { "field": "os_sph", ... },
    { "field": "od_dt", "label": "OD Distance PD", "required": true, ... },
    { "field": "os_dt", ... },
    { "field": "od_seg_hd", "label": "OD Fitting Height", "required": false, ... },
    { "field": "od_bvd", "label": "OD Back Vertex Distance", "required": false, ... },
    { "field": "os_bvd", ... },
    { "field": "size_lens_width", "label": "Frame Eye Size (A)", "required": true, ... },
    { "field": "b_value", "label": "Frame B", "required": true, ... },
    { "field": "size_bridge_width", "label": "Frame DBL", "required": true, ... },
    { "field": "panto", "label": "Pantoscopic Angle", "required": false, ... },
    { "field": "wrap_angle", "label": "Frame Bow Angle", "required": false, ... },
    { "field": "lab_instructions", "label": "Special Instructions", "required": false, ... }
  ]
}
```

### Логика фронта:
```
can_order === false                          → "Manual Order", без кнопки
can_order && provider === "zeiss"
  && zeiss_auth.filled === false             → показать "Login to Zeiss"
can_order && ready === false                 → Order (disabled) + незаполненные поля
can_order && ready === true                  → Order (active)
```

---

## 2. Отправка заказа

```
POST /api/ticket/{ticket_id}/order
```
Body не нужен.

| HTTP | Когда |
|---|---|
| `200` | Заказ отправлен успешно |
| `400` | Custom lens — `"manual orders cannot be submitted electronically"` |
| `409` | Уже отправлен — возвращает order_id + текущий статус |
| `422` | Валидация не прошла — список полей |
| `501` | Zeiss order ещё не реализован |

### 200 — успех:
```json
{ "vw_order_id": "VW123456", "status": "Sent", "error_list": "" }
```

### 409 — повторная отправка:
```json
{ "error": "order already submitted...", "vw_order_id": "VW123456", "status": "In Production" }
```

### 422 — не хватает данных:
```json
{ "error": "validation failed", "errors": [{ "field": "od_sph", "message": "OD Sphere is required" }] }
```
