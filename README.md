# Микросервис для работы с балансом пользователей

**Задача:**

Необходимо реализовать микросервис для работы с балансом пользователей (зачисление средств, списание средств, перевод средств от пользователя к пользователю, а также метод получения баланса пользователя). Сервис должен предоставлять HTTP API и принимать/отдавать запросы/ответы в формате JSON.

**Основное задание**

Метод начисления средств на баланс. Принимает id пользователя и сколько средств зачислить.
Метод резервирования средств с основного баланса на отдельном счете. Принимает id пользователя, ИД услуги, ИД заказа, стоимость.
Метод признания выручки – списывает из резерва деньги, добавляет данные в отчет для бухгалтерии. Принимает id пользователя, ИД услуги, ИД заказа, сумму.
Метод получения баланса пользователя. Принимает id пользователя.

---
**Диаграмма классов**

![Диаграмма классов](img/Diagram.png)

---
**Решенные задачи**

1. Метод начисления средств на баланс. Принимает id пользователя и сколько средств зачислить.
2. Метод резервирования средств с основного баланса на отдельном счете. Принимает id пользователя, ИД услуги, ИД заказа, стоимость.
3. Метод признания выручки – списывает из резерва деньги, добавляет данные в отчет для бухгалтерии. Принимает id пользователя, ИД услуги, ИД заказа, сумму.
4. Метод получения баланса пользователя. Принимает id пользователя.

**Дополнительно**
- Получение списка транзакций с комментариями
- Добавленна обработка ошибок с возвратом соответствующего типа и json 

---
**Запуск**

После клонирования репозитория достаточно выполнить `make` в папке проекта
> Для завершения: `make clean`.
> Для перезапуска: `make restart`.
> Для генерации swagger `make generate_swagger`.
---
**Запросы**

- *Получение баланса пользователя*
```
curl -X 'GET' \
  'http://localhost:8080/api/v1/balances/user1' \
  -H 'accept: application/json'
```

- *Пополнение баланса пользователя (создание нового с нужным балансом)*
```
curl -X 'POST' \
  'http://localhost:8080/api/v1/balances/{user_id}/credit' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "currency": 30
}'
```

- *Резервирование средств пользователя*
```
curl -X 'POST' \
  'http://localhost:8080/api/v1/balances/{user_id}/reserve' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "currency": 50,
  "order_id": "1",
  "service_id": "shop"
}'
```

- *Признание выручки*
```
curl -X 'POST' \
  'http://localhost:8080/api/v1/balances/{user_id}/commit' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "currency": 50,
  "order_id": "1",
  "service_id": "shop"
}'
```

- *Составление отчета*

Просто все операции по пользователю:
```
curl -X 'GET' \
  'http://localhost:8080/api/v1/balances/{user_id}/operations' \
  -H 'accept: application/json'
```
Операции с доп. параметрами (limit - кол-во выведенных, offset - отступ от первой записи, order_by - сортировка по стоимости/времени, desc - порядок)
```
curl -X 'GET' \
  'http://localhost:8080/api/v1/balances/{user_id}/operations?limit=10&offset=0&order_by=date&desc=true' \
  -H 'accept: application/json'
```

---
**Комментарии**

1. Так как резерв баланса не используется ни в какой сторонней бизнес-логике, было принято решение вынести его из сущности и хранить только в базе данных.
2. Формат, в котором хранятся деньги на данный момент имеет тип int, обернутый в структурку. Предполагается, что суммы изначально указаны в копейках.Возможна его замена на decimal или же на два int`а, которые будут хранить условные рубли и копейки

**Update**
1. Резерв перемещен в таблицу с операциями
2. Теперь можно списать резерв с операции частично
3. Добавлен swagger. Swager UI по адресу `http://localhost:8080/swagger/`
4. Убран метод rollback, так как его реализация значительно усложняется в связи с изменениями в резерве и подтверждении выручки
5. Диаграмма классов поменялась
