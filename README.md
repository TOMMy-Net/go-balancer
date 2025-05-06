# go-balancer

* Перед запуском проекта необходимо заполнить config.yaml:

```

server-port: 8000
api-port: 8080
backends:
  endpoints: 
    -"localhost:4000"
    -"localhost:5000"
  health-interval: 5
rate-limiter:
  default-interval: 5
  default-capacity: 100
  default-refill-rate: 10

database: "postgres://postgres:770948@localhost:5432/loadbalancer?sslmode=disable&TimeZone=UTC"
```

* Создать файл load_balancer.log
* Запуск проекта:

```
docker-compose up
```

# Улучшения

Разделить API и балансировщик на два разных контейнера и добавить брокер сообщений (beanstalkd, rabbitMQ) для общения и синхронизации данных

# Пожелания

Дается мало времени так как прислали на майских и многие не дома в это время(
