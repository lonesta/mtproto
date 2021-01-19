# MTProto

![help wanted](https://img.shields.io/badge/-help%20wanted-success)
[![godoc reference](https://pkg.go.dev/badge/github.com/lonesta/mtproto?status.svg)](https://pkg.go.dev/github.com/lonesta/mtproto)
[![Go Report Card](https://goreportcard.com/badge/github.com/lonesta/mtproto)](https://goreportcard.com/report/github.com/lonesta/mtproto)
[![license MIT](https://img.shields.io/badge/license-MIT-green)](https://github.com/lonesta/mtproto/blob/master/README.md)
[![chat telegram](https://img.shields.io/badge/chat-telegram-0088cc)](https://bit.ly/2xlsVsQ)
![version v0.1.0](https://img.shields.io/badge/version-v0.1.0-red)
![unstable](https://img.shields.io/badge/stability-unstable-yellow)
<!--
code quality
golangci
contributors
go version
gitlab pipelines
-->

![FINALLY!](/docs/assets/finally.jpg) Полностью нативная имплементация MTProto на Golang!


[english](https://github.com/lonesta/mtproto/blob/master/docs/en_US/README.md) **русский**

<p align="center">
<img src="/docs/assets/MuffinMan-AgADRAADO2AkFA.gif"/>
</p>

## <p align="center">Фичи</p>

<div align="right">
<h3>Полностью нативная реализация</h3>
<img src="/docs/assets/ezgif-3-a6bd45965060.gif" align="right"/>
Вся библиотека начиная с отправки запросов и шифрования и заканчивая сериализацией шифровния написаны исключительно на golang. Для работы с библиотекой не требуется никаких лишних зависимостей.
</br></br></br></br></br></br>
</div>

<div align="left">
<h3>Самая свежая версия API (117+)</h3>
<img src="/docs/assets/ezgif-3-19ced73bc71f.gif" align="left"/>
Реализована поддержка всех возможностей API Telegram и MTProto включая функцию видеозвонков и комментариев к постам. Вы можете сделать дополнительный pull request на обновление данных!
</br></br></br></br></br></br></br>
</div>

<div align="right">
<h3>Реактивные обновления (сгенерировано из TL спецификаций)</h3>
<img src="/docs/assets/ezgif-3-5b6a808d2774.gif" align="right"/>
Все изменения в клиентах TDLib и Android мониторятся на предмет появления новых фич и изменений в TL схемах. Новые методы и объекты появляются просто по добавлению новых строк в схеме и обновления сгенерированного кода!
</br></br></br></br></br>
</div>

<div align="left">
<h3>Implements ONLY network tools</h3>
<img src="/docs/assets/ezgif-3-3ac8a3ea5713.gif" align="left"/>
Никаких SQLite баз данных и кеширования ненужных <b>вам</b> файлов. Вы можете использовать только тот функционал, который вам нужен. Вы так же можете управлять способом сохранения сессий, процессом авторизации, буквально всем, что вам необходимо!
</br></br></br></br></br>
</div>

<div align="right">
<h3>Multiaccounting, Gateway mode</h3>
<img src="/docs/assets/ezgif-3-7bcf6dc78388.gif" align="right"/>
Вы можете использовать больше 10 аккаунтов одновременно! xelaj/MTProto не создает большого оверхеда по вычислительным ресурсам, поэтому вы можете иметь огромное количество инстансов соединений и не переживать за перерасход памяти!
</br></br></br></br></br>
</div>

## How to use

<!--
**СЮДА ИЗ asciinema ЗАПИХНУТЬ ДЕМОНСТРАЦИЮ**
![preview]({{ .PreviewUrl }})
-->

MTProto очень сложен в реализации, но при этом очень прост в использовании. По сути вы общаетесь с серверами Telegram посредством отправки сериализованых структур (аналог gRPC, разработанный Telegram llc.). Выглядит это примерно так:

```go
func main() {
    client := &Telegram.NewClient()
    // for each method there is specific struct for serialization (<method_name>Params{})
    result, err := client.MakeRequest(&telegram.GetSomeInfoParams{FromChatId: 12345})
    if err != nil {
        panic(err)
    }

    resp, ok := result.(*SomeResponseObject)
    if !ok {
        panic("Oh no! Wrong type!")
    }
}
```

Однако, есть более простой способ отправить запрос, который уже записан в TL спецификации API:

```go
func main() {
    client := &Telegram.NewClient()
    resp, err := client.GetSomeInfo(12345)
    if err != nil {
        panic(err)
    }

    // resp will be already asserted as described in TL specs of API
    // if _, ok := resp.(*SomeResponseObject); !ok {
    //     panic("No way, we found a bug! Create new issue!")
    // }

    println(resp.InfoAboutSomething)
}
```

Вам не стоит задумываться о реализации шифрования, обмена ключами, сохранении и восстановлении сессии, все уже сделано за вас.

**Code examples are [here](https://github.com/lonesta/mtproto/blob/master/examples)**

**Full docs are [here](https://pkg.go.dev/github.com/lonesta/mtproto)**

## Getting started

### Simple How-To

Все как обычно, вам необходимо загрузить пакет с помощью `go get`:

``` bash
go get github.com/lonesta/mtproto
```

Далее по желанию вы можете заново сгенерировать исходники структур методов и функций, для этого используйте команду `go generate`

``` bash
go generate github.com/lonesta/mtproto
```

Все! Больше ничего и не надо!

### что за InvokeWithLayer?

Это специфическая особенность Telegram, для создания соединения и получения информации о текущей конфигурации серверов, нужно сделать что-то подобное:

```go
    resp, err := client.InvokeWithLayer(apiVersion, &telegram.InitConnectionParams{
        ApiID:          124100,
        DeviceModel:    "Unknown",
        SystemVersion:  "linux/amd64",
        AppVersion:     "0.1.0",
        SystemLangCode: "en",
        LangCode:       "en",
        Proxy:          nil,
        Params:         nil,
        // HelpGetConfig() is ACTUAL request, but wrapped in IvokeWithLayer
        Query:          &telegram.HelpGetConfigParams{},
    })
```

### Как произвести авторизацию по телефону?

**Пример [здесь](https://github.com/lonesta/mtproto/blob/master/examples/auth)**

```go
func AuthByPhone() {
    resp, err := client.AuthSendCode(
        yourPhone,
        appID,
        appHash,
        &telegram.CodeSettings{},
    )
	if err != nil {
        panic(err)
    }

    // Можно выбрать любой удобный вам способ ввода,
    // базовые параметры сессии можно сохранить в любом месте
	fmt.Print("Auth code:")
	code, _ := bufio.NewReader(os.Stdin).ReadString('\n')
    code = strings.Replace(code, "\n", "", -1)

    // это весь процесс авторизации!
    fmt.Println(client.AuthSignIn(yourPhone, resp.PhoneCodeHash, code))
}
```

Все! вам не требуется никаких циклов или чего-то подобного, код уже готов к асинхронному выполнению, вам нужно только выполнить действия прописанные в документации к Telegram API

### Документация пустует! Почему?

Объем документации невероятно огромен. Мы бы готовы задокументировать каждый метод и объект, но это огромное количество работы. Несмотря на это, **все** методы **уже** описаны [здесь](https://core.telegram.org/methods), вы можете так же спокойно их

### Работает ли этот проект под Windows?

Технически — да. Компоненты не были заточены под определенную архитектуру. Однако, возможности протестировать у разработчиков не было. Если у вас возникли проблемы, напишите в issues, мы постараемся помочь

## Who use it

## Contributing

Please read [contributing guide](https://github.com/lonesta/mtproto/blob/master/doc/en_US/CONTRIBUTING.md) if you want to help. And the help is very necessary!

## TODO

[ ]

## Authors

* **Richard Cooper** <[rcooper.xelaj@protonmail.com](mailto:rcooper.xelaj@protonmail.com)>

## License

<b style="color:red">WARNING!</b> This project is only maintained by Xelaj inc., however copyright of this source code **IS NOT** owned by Xelaj inc. at all. If you want to connect with code owners, write mail to <a href="mailto:up@khsfilms.ru">this email</a>. For all other questions like any issues, PRs, questions, etc. Use GitHub issues, or find email on official website.

This project is licensed under the MIT License - see the [LICENSE](https://github.com/lonesta/mtproto/blob/master/doc/en_US/LICENSE.md) file for details