# Seismo

Система для сбора, обработки и анализа данных о сейсмической активности.

Проект создан как заготовка для практического обучения в рамках курса преподавания языка Go, и предполагает дальнейшее  развитие учащимися.


## Назначение и краткое описание

Seismo задумана как система сервисов и баз данных, которая способна собирать сообщения о сейсмической активности из различных открытых источников, сохранять их, уточнять, фильтровать, и создавать на их основе массивы данных, формирующие картину сейсмической активности на больших территориях нашей планеты.
Кроме того, сервисы Seismo должны обеспечивать доступ пользователей к данным как через API, так и посредством пользовательского веб-интерфейса, а так же предоставлять инструменты для анализа данных.

В качестве проекта-заготовки в данном репозитории предлагается простой сервис сбора и сохранения сообщений о сейсмической активности, а также необходимые для его работы пакеты. Кроме того, имеется простой инструмент для извлечения и сохранения сообщений для одного конкретного источника.

Проект имеет в первую очередь обучающую направленность, и призван познакомить студентов с некоторыми основными принципами и шаблонами разработки на языке Go.  

Предлагается использование представленного в данном репозитории проекта-заготовки как основы для развития системы в намеченном выше направлении силами студентов-энтузиастов, а так же всех желающих это делать, каковы бы ни были их конечные цели. Главное, чтобы цели эти были добрыми. =)

Распространяется под лицензией MIT.


## Архитектура и компоненты
Для реализации функционала в составе Seismo предлагается создать описанные ниже компоненты.
Потенциально, при дальнейшем развитии, сервисы Seismo могут быть доступны для равёртывания в контейнерах и с помощью систем оркестрации, таких, как Kubernetes.  

### Collector
Collector - сервис, способный одновременно "прослушивать" несколько разнотипных источников сообщений, получать из них сообщения в стандартном виде, и сохранять их в базу данных, или несколько баз данных разного типа. Кроме того, сервис должен иметь возможность возобновлять прослушивание источников, если оно прервалось.

Для работы с источниками разного типа, Collector должен иметь соответствующий абстрактный интерфейс, конкретная реализация которого зависит от конкретного типа источника.

То же касается и возможности взаимодействия с различными типами баз данных, т.е. есть необходимость в интерфейсе и его различных реализациях для разных типов СУБД.    

### CollectorDb
Collector Database - база данных для сохранения полученных сообщений Collector'ом и их последующего извлечения.
Сообщения имеют достаточно простую структуру, которая может меняться, несколько сообщений могут описывать одно и тоже сейсмическое событие, сообщения могут дублироваться по самым разным причинам. Учитывая простую и изменчивую структуру данных и примитивный уровень предполагаемых запросов, для этих целей может использоваться документо-ориентированная база данных, например, MongoDb. Учитывая промежуточное положение такой базы данных, вместо неё можно также использовать брокер сообщений.

### DataComposer
DataComposer - сервис, который периодически извлекает сообщения из CollectorDb, устраняет дублирование сообщений, выявляет разные сообщения об одном и том же сейсмическом событии, определяет наиболее качественные данные, фильтрует их, и сохраняет в подготовленном для дальнейшего использования в аналитических целях виде в отдельную базу данных SeismoDb на постоянное хранение.

DataComposer также как Collector должен допускать возможность работы с разными СУБД.

Учитывая, что сохранение таких данных удобнее осуществлять крупными блоками, возможно имеет смысл предусмотреть некоторое промежуточное оперативное хранилище, накапливающее эти данные порциями, перед "сбросом" их в SeismoDb. Это особенно актуально, если для реализации SeismoDb будет выбрана колоночная СУБД.

### SeismoDb
SeismoDb - база данных для хранения "сведённой" и обработанной статистической информации о сесмической активности. Данные не предполагают частого последующего изменения, поэтому, в случае, если реализация системы не предполагает доступности новых данных пользователям  в режиме реального времени, т.е. анализ данных делается значительно позже их появления (спустя часы), имеет смысл применить для реализации колоночную СУБД, например, ClickHouse. Если же необходим оперативный доступ пользователей к самым свежим, только что полученным данным, стоит использовать табличную реляционную СУБД.
Также можно создать комбинацию из табличного и колоночного хранилища, первое из которых будет отвечать за меньшую оперативно пополняемую часть, а второе - за основной постоянно хранимый массив, и периодически пополняться крупными блоками данных из первого. Такое решение может серьёзно усложнить взаимодействие сервисов с базами данных, но при этом решить проблемы как со скоростью доступа, так и со скоростью записи.

### Distributor
Distributor - сервис отвечающий за выдачу данных из SeismoDb. Предоставляет API, через который Gateway запрашивает у него данные для пользователей. Учитывая специфику данных (статистические), возможным выбором может быть gRPC.   

### Gateway
Gateway - интеллектуальный шлюз, который отвечает за регистрацию, авторизацию и аутентификацию пользователей, и через который осуществляется взаимодействие с системой извне посредством API. Gateway хранит информацию о пользователях в базе данных AuthDb. Также получает данные для пользователя от сервиса Distributor, обращаясь к нему по API.

### AuthDb
AuthDb - база данных для хранения информации об авторизации и аутентификации. Наиболее подходящей здесь является реляционная табличная СУБД, например PostrgeSQL.

### WebApp
WebApp - веб-приложение, обеспечивающее графический пользовательский интерфейс. Авторизацию и аутентификацию пользователей осуществляет обращаясь к соответствующему API сервиса Gateway. Имеет прямо доступ к основному хранилищу SeismoDb.

### LogDb
Т.к. в дальнейшем предполагается развертывание сервисов системы в контейнерах, необходимо предусмотреть отдельную базу данных для логирования. 

## Текущее состояние (что реализовано)

### Collector 
В настоящий момент создан простой работающий сервис сбора сообщений, способный "прослушивать" несколько источников сообщений одновременно и сохранять данные в БД. Также способен перезапускать получение сообщений с момента разрыва соединения. Представлен пакетом seismo/collector и пакетом main. Collector использует пакет provider, и его внутренние пакеты для работы с источниками сообщений. Настройки сервис считывает при запуске из конфигурационного файла, полный путь к которому может быть передан как значение флага команды, либо получен из переменной окружения. 

### seismo/collector/db
Пакет seismo/collector/db обеспечивает основные типы (в том числе интерфейс Adapter) для взаимодействия с различными СУБД. Кроме того, предоставляет фабричную функцию, локализующую создание экземпляра конкретной реализации интерфейса Adapter, в зависимости от передаваемых в функцию настроек базы данных.

### seismo/collector/db/mongodb
Пакет seismo/collector/db/mongodb предоставляет инструменты для взаимодействия Collector'а с MongoDb, реализует интерфейс provider.Adapter.

### seismo/collector/db/stubdb
Пакет seismo/collector/db/stubdb предоставляет фиктивную реализацию интерфейса provider.Adapter, имитирующую взаимодействие с базой данных. Может использоваться в тестовых целях как "заглушка" для интерфейса.

### seismo/provider
Пакет seismo/provider содержит основные типы, такие, как тип сообщения, а также интерфейсы, необходимые для реализации работы с разными источниками сейсмических сообщений, в первую очередь интерфейс Watcher. 

### seismo/provider/seishub
Пакет seismo/provider/seishub предоставляет большой набор инструментов для работы с конкретным источником сообщений - SEISHUB'ом, представляемым Алтае-Саянским филиалом ФИЦ ЕГС РАН, а так же реализацию интерфейса provider.Wahcher. 

### seishub-util
Простое консольное приложение, позволяющее работать с источником SEISHUB, извлекать из него и сохранять сообщения в виде файлов. Написано для вспомогательных целей. 

### seismo/provider/pseudo
Пакет seismo/provider/pseudo предоставляет локальный источник фиктивных сообщений о сейсмических событиях, реализуя интерфейс provider.Watcher. Сообщения создаются случайным образом через заданный промежуток времени. Используется в тестовых целях.

### seismo/provider/crt
Пакет seismo/provider/crt локализует фабричные функции для создания экземпляров, реализующих абстракции пакета seismo/provider. В настоящее время такая фабричная функция одна - NewWatcher, создающая экземпляр конкретной реализации интерфейса provider.Watcher, в зависимости от передаваемых в функцию настроек. Также пакет обеспечивает дополнительный слой, позволяющий избежать циклических зависимостей между пакетам seismo/provider и его внутренними пакетами.

### CollectorDb
Т.к. MongoDb по умолчанию создаёт необходимые структуры при сохранении данных, в настоящее время в проекте нет специального кода для создания базы и её внутренних структур. 


## На что обратить внимание

Учащиеся могут ознакомиться в коде с некоторыми приёмами разработки в целом и на языке Go в частности.
Ниже представлен список того, на что стоит обратить внимание.

### Общие принципы и шаблоны разработки, иллюстрируемые кодом проекта
1. Принцип использования абстракции для обеспечения независимости от конкретной реализации представлен интерфейсом Watcher в пакете seismo/provider, и его реализациями в пакетах seismo/provider/seishub и seismo/provider/pseudo, непосредственное же их применение можно увидеть в пакете seismo/collector в функциях CreateWatchers и RestartWatchers, а также в функции main самого сервиса Collector. Также данный принцип иллюстрируется в пакете seismo/collector/db интерфейсом Adpater.

2. Принципы инверсии зависимостей и локализации создания экземпляров конкретных реализаций также обеспечиваются перечисленными в предыдущем пункте структурами. В основной логике сервиса Collector зависимости от конкретных особенностей источников удаётся избежать благодаря промежуточному слою provider, от которого зависят сами конкретные реализации (т.е., зависимость инвертируется, меняет своё направление). Непосредственно за локализацию создания отвечает пакет seismo/provider/crt, содержащий фабричные функции. Кроме того, роль "инвертирующего" промежуточного слоя между логикой Collector'а и конкретными СУБД играет пакет seismo/collector/db. 

3. Циклических зависимостей между пакетом seismo/provider и его внутренними пакетами-реализациями (например, seismo/provider/seishub) удаётся избежать путём выделения фабричных функций в отдельный пакет seismo/provider/crt. Обратите внимание, что фабричные функции пакета seismo/collector/db не выделены в отдельный пакет, так как не создают циклических зависимостей. Несмотря на то, что избегание циклических зависимостей - характерная черта языка Go, по возможности избегать подобного явления рекомендуется при программировании вообще.

4. Шаблон проектирования "Состояние" реализуется структурами Hub в пакетах seismo/provider/seishub и seismo/provider/pseudo. Несмотря на то, что состояний в настоящее время только два, а методов, зависящих от состояния не много, использование шаблона оправдано, т.к. позволяет избежать лишних условных конструкций в методах, и будет особенно полезно при развитии и усложнении поведения структур. 

### Приёмы программирования, характерные для языка Go.
1. Шаблон for-select. Примеры простого использования шаблона for-select можно увидеть в методе Hub.generateMessages пакета seismo/provider/pseudo, методов watch и getStartMsgNum пакета seismo/provider/seishub а также в функции main сервиса Collector.

2. Отмена. Обратите внимание, что во всех перечисленных в предыдущем пункте случаях в конструкции select применяется отмена через контекст.

3. Fan-out. Вариант реализации схемы fan-out, т.е., асинхронной обработки несколькими горутинами данных, получаемых из одного канала, можно наблюдать в методе Hub.Extract пакета seismo/provider/seishub. Обратите внимание, что здесь применяется схема с фиксированным числом заранее запущенных горутин. 

4. WaitGroup. Использование структуры sync.WaitGroup для ожидания завершения запущенных горутин можно наблюдать в том же методе Hub.Extract пакета seismo/provider/seishub.

5. Канал как возвращаемое значение. Обратите внимание, что метод StartWatch интерфейса Watcher в пакете seismo/provider имеет канал как возвращаемое значение. С реализациями этого примёма можно ознакомиться в пакетах seismo/provider/seishub и seismo/provider/pseudo.

6. Канал каналов. В пакете seismo/collector функция RestartWatchers использует канал, для передачи новых открытых каналов функции и MergeWatchPipes.

7. Объединение каналов. Функция MergeWatchPipes пакета seismo/collector реализует перенаправление данных из нескольких каналов в один. 


## Обучающие задания

1. Система логирования. В предлагаемом проекте-заготовке используется стандартный логгер языка Go. Учащимся предлагается создать специальную базу данных LogDb для хранения логов, а также реализовать и применить свой собственный логгер, или воспользоваться существующими решениями.

2. Тесты. Код проекта-заготовки имеет недостаточное количество тестов. Учащимся предлагается добавить необходимые тесты, используя, среди прочего, пакет httptest.

3. Провайдеры для различных источников сообщений. Учащимся предлагается создать реализации интерфейса provider.Watcher для различных источников сообщений о сейсмических событиях. Выполнить их как отдельные пакеты, вложенные в пакет seismo/provider.

4. DataComposer и SeismoDb. Создать сервис DataComposer и базу данных SeismoDb согласно принципам, изложенным в разделе "Архитектура и компоненты".

5. Distributor. Реализовать сервис Distributor, разработав API согласно описанию в разделе "Архитектура и компоненты".

6. Gateway и AuthDb. Разработать отвечающий за регистрацию, авторизацию и аутентификацию шлюз и его базу данных, согласно описанию в разделе "Архитектура и компоненты".




