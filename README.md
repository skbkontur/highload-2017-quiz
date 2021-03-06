# Быстрая фильтрация метрик из Графита

Привет, участник HighLoad++. Создатели [Moira 2.0](https://github.com/moira-alert) приготовили для тебя задание.

*Moira — система для отправки уведомлений о нештатных ситуациях. Приложения отправляют в Graphite десятки тысяч метрик в секунду, а Moira получает их полную копию. Уведомления в Moira настроены только на несколько процентов метрик. Поэтому важно уметь быстро фильтровать метрики по известным шаблонам.*

**Задание.** Есть [тривиальная, но не очень быстрая](https://github.com/skbkontur/highload-2017-quiz/blob/master/matcher.go) реализация фильтра в файле `matcher.go`. Она проходит [модульные тесты](https://github.com/skbkontur/highload-2017-quiz/blob/master/matcher_test.go) на [Travis CI](https://travis-ci.org/skbkontur/highload-2017-quiz/jobs/298133612). Напишите свою, более быструю реализацию фильтра и пришлите пулл-реквест.

**Условия:**
* Один участник может отправить только один пулл-реквест.
* В пулл-реквесте можно менять только файл [`fastmatcher.go`](https://github.com/skbkontur/highload-2017-quiz/blob/master/fastmatcher.go).
* В пулл-реквесте должны проходить модульные тесты.
* Из-за погрешности в работе Travis CI, реализации с разницей в скорости менее 2 % считаются имеющими равную скорость.

**Пример.** Алексей Кирпичников прислал [пулл-реквест](https://github.com/skbkontur/highload-2017-quiz/pull/1) с другой реализацией фильтра. Она работает на 11 % быстрее, как утверждает [Travis CI](https://travis-ci.org/skbkontur/highload-2017-quiz/builds/298137816).

**Победители и награждение:**
* Победителями считаются 10 участников с самыми производительными реализациями.
* Приём пулл-реквестов заканчивается 8 ноября 2017 в 17:00 (МСК).
* В 17:10 [Алексей Кирпичников](https://github.com/beevee/), один из создателей Moira, прокомментирует самые производительные решения и наградит победителей. У нас есть книжки, тёплые варежки и термокружки от [СКБ Контур](https://github.com/skbkontur).
