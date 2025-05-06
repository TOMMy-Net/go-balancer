# go-balancer

* Перед запуском проекта необходимо заполнить config.yaml:

```
server-port: 8000
api-port: 8080
backends:
  endpoints: 
    - "localhost:4000"
    - "localhost:5000"
  health-interval: 5
rate-limiter:
  default-interval: 5
  default-capacity: 100
  default-refill-rate: 10

database: 
  host: "pg_db"
  user: "postgres"
  password: "770948"
  name_db: "loadbalancer"
  port: "5432"
  ssl: "disable"
```

* Запуск проекта:

```
docker-compose up
```

# Описание

Балансировщик работает на одном порту, а доступ к api балансировщика на другом.

# Улучшения

Разделить API и балансировщик на два разных контейнера и добавить брокер сообщений (beanstalkd, rabbitMQ) для общения и синхронизации данных

# Пожелания

Дается мало времени так как прислали на майских и многие не дома в это время(
