# {{ .Project.Name }}

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


[english](https://{{ .PackageUrl }}/blob/master/doc/en_US/README.md) **русский**

{{ .Title.Text }}

<p align="center">
<img src="{{ .Title.ImageUrl }}"/>
</p>

## Как установить

TODO

## Как использовать

![preview]({{ .PreviewUrl }})

**Примеры кода [здесь](https://{{ .PackageUrl }}/blob/master/examples)**

### Simple How-To

{{ .AdditionalHowto }}

{{ .SimpleFAQ }}

## Вклад в проект

пожалуйста, прочитайте [информацию о помощи]https://{{ .PackageUrl }}/blob/master/doc/ru_RU/CONTRIBUTING.md), если хотите помочь. А помощь очень нужна!

## TODO

{{ range $item := .TODO }}* {{ $item }}
{{ end }}
## Авторы

{{ range $author := .Authors }}* **{{ $author.Name }}** — [{{ $author.Nick }}](https://github.com/{{ $author.Nick }})
{{ end }}
## Лицензия

This project is licensed under the MIT License - see the [LICENSE](https://{{ .PackageUrl }}/blob/master/doc/ru_RU/LICENSE.md) file for details
