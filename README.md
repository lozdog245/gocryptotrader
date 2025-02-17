<img src="https://github.com/thrasher-corp/gocryptotrader/blob/master/web/src/assets/page-logo.png?raw=true" width="350px" height="350px" hspace="70">

[![Build Status](https://travis-ci.com/thrasher-corp/gocryptotrader.svg?branch=master)](https://travis-ci.com/thrasher-corp/gocryptotrader)
[![Software License](https://img.shields.io/badge/License-MIT-orange.svg?style=flat-square)](https://github.com/thrasher-corp/gocryptotrader/blob/master/LICENSE)
[![GoDoc](https://godoc.org/github.com/thrasher-corp/gocryptotrader?status.svg)](https://godoc.org/github.com/thrasher-corp/gocryptotrader)
[![Coverage Status](http://codecov.io/github/thrasher-corp/gocryptotrader/coverage.svg?branch=master)](http://codecov.io/github/thrasher-corp/gocryptotrader?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/thrasher-corp/gocryptotrader)](https://goreportcard.com/report/github.com/thrasher-corp/gocryptotrader)

A cryptocurrency trading bot supporting multiple exchanges written in Golang.

**Please note that this bot is under development and is not ready for production!**

## Community

Join our slack to discuss all things related to GoCryptoTrader! [GoCryptoTrader Slack](https://join.slack.com/t/gocryptotrader/shared_invite/enQtNTQ5NDAxMjA2Mjc5LTc5ZDE1ZTNiOGM3ZGMyMmY1NTAxYWZhODE0MWM5N2JlZDk1NDU0YTViYzk4NTk3OTRiMDQzNGQ1YTc4YmRlMTk)

## Exchange Support Table

| Exchange | REST API | Streaming API | FIX API |
|----------|------|-----------|-----|
| Alphapoint | Yes  | Yes        | NA  |
| ANXPRO | Yes  | No        | NA  |
| Binance| Yes  | Yes        | NA  |
| Bitfinex | Yes  | Yes        | NA  |
| Bitflyer | Yes  | No      | NA  |
| Bithumb | Yes  | NA       | NA  |
| BitMEX | Yes | Yes | NA |
| Bitstamp | Yes  | Yes       | No  |
| Bittrex | Yes | No | NA |
| BTCMarkets | Yes | No       | NA  |
| BTSE | Yes | Yes | NA |
| COINUT | Yes | Yes | NA |
| Exmo | Yes | NA | NA |
| CoinbasePro | Yes | Yes | No|
| GateIO | Yes | Yes | NA |
| Gemini | Yes | Yes | No |
| HitBTC | Yes | Yes | No |
| Huobi.Pro | Yes | Yes | NA |
| ItBit | Yes | NA | No |
| Kraken | Yes | Yes | NA |
| Lbank | Yes | No | NA |
| LakeBTC | Yes | No | NA |
| LocalBitcoins | Yes | NA | NA |
| OKCoin International | Yes | Yes | No |
| OKEX | Yes | Yes | No |
| Poloniex | Yes | Yes | NA |
| Yobit | Yes | NA | NA |
| ZB.COM | Yes | Yes | NA |

We are aiming to support the top 20 highest volume exchanges based off the [CoinMarketCap exchange data](https://coinmarketcap.com/exchanges/volume/24-hour/).

** NA means not applicable as the Exchange does not support the feature.

## Current Features

+ Support for all Exchange fiat and digital currencies, with the ability to individually toggle them on/off.
+ AES256 encrypted config file.
+ REST API support for all exchanges.
+ Websocket support for applicable exchanges.
+ Ability to turn off/on certain exchanges.
+ Ability to adjust manual polling timer for exchanges.
+ Communication packages (Slack, SMS via SMSGlobal, Telegram and SMTP)
+ HTTP rate limiter package.
+ Forex currency converter packages (CurrencyConverterAPI, CurrencyLayer, Fixer.io, OpenExchangeRates)
+ Packages for handling currency pairs, tickers and orderbooks.
+ Portfolio management tool; fetches balances from supported exchanges and allows for custom address tracking.
+ Basic event trigger system.
+ WebGUI.

## Planned Features

Planned features can be found on our [community Trello page](https://trello.com/b/ZAhMhpOy/gocryptotrader).

## Contribution

Please feel free to submit any pull requests or suggest any desired features to be added.

When submitting a PR, please abide by our coding guidelines:

+ Code must adhere to the official Go [formatting](https://golang.org/doc/effective_go.html#formatting) guidelines (i.e. uses [gofmt](https://golang.org/cmd/gofmt/)).
+ Code must be documented adhering to the official Go [commentary](https://golang.org/doc/effective_go.html#commentary) guidelines.
+ Code must adhere to our [coding style](https://github.com/thrasher-corp/gocryptotrader/blob/master/.github/CONTRIBUTING.md).
+ Pull requests need to be based on and opened against the `master` branch.

## Compiling instructions

Download and install Go from [Go Downloads](https://golang.org/dl/) for your
platform.

### Linux/OSX

GoCryptoTrader is built using [Go Modules](https://github.com/golang/go/wiki/Modules) and requires Go 1.11 or above
Using Go Modules you now clone this repository **outside** your GOPATH

```bash
git clone https://github.com/thrasher-corp/gocryptotrader.git
cd gocryptotrader
go build
mkdir ~/.gocryptotrader
cp config_example.json ~/.gocryptotrader/config.json
```

### Windows

```bash
git clone https://github.com/thrasher-corp/gocryptotrader.git
cd gocryptotrader
go build
copy config_example.json %APPDATA%\GoCryptoTrader\config.json
```

+ Make any neccessary changes to the `config.json` file.
+ Run the `gocryptotrader` binary file inside your GOPATH bin folder.

## Donations

<img src="https://github.com/thrasher-corp/gocryptotrader/blob/master/web/src/assets/donate.png?raw=true" hspace="70">

If this framework helped you in any way, or you would like to support the developers working on it, please donate Bitcoin to:

***1F5zVDgNjorJ51oGebSvNCrSAHpwGkUdDB***

## Binaries

Binaries will be published once the codebase reaches a stable condition.

## Contributor List

### A very special thank you to all who have contributed to this program:

|User|Github|Contribution Amount|
|--|--|--|
| thrasher- | https://github.com/thrasher- | 548 |
| shazbert | https://github.com/shazbert | 176 |
| gloriousCode | https://github.com/gloriousCode | 155 |
| xtda | https://github.com/xtda | 18 |
| ermalguni | https://github.com/ermalguni | 14 |
| vadimzhukck | https://github.com/vadimzhukck | 10 |
| 140am | https://github.com/140am | 8 |
| marcofranssen | https://github.com/marcofranssen | 8 |
| cranktakular | https://github.com/cranktakular | 5 |
| MadCozBadd | https://github.com/MadCozBadd | 3 |
| leilaes | https://github.com/leilaes | 3 |
| crackcomm | https://github.com/crackcomm | 3 |
| andreygrehov | https://github.com/andreygrehov | 2 |
| bretep | https://github.com/bretep | 2 |
| woshidama323 | https://github.com/woshidama323 | 2 |
| gam-phon | https://github.com/gam-phon | 2 |
| cornelk | https://github.com/cornelk | 2 |
| if1live | https://github.com/if1live | 2 |
| soxipy | https://github.com/soxipy | 2 |
| herenow | https://github.com/herenow | 2 |
| blombard | https://github.com/blombard | 1 |
| CodeLingoBot | https://github.com/CodeLingoBot | 1 |
| CodeLingoTeam | https://github.com/CodeLingoTeam | 1 |
| Daanikus | https://github.com/Daanikus | 1 |
| daniel-cohen | https://github.com/daniel-cohen | 1 |
| DirectX | https://github.com/DirectX | 1 |
| frankzougc | https://github.com/frankzougc | 1 |
| starit | https://github.com/starit | 1 |
| Jimexist | https://github.com/Jimexist | 1 |
| lookfirst | https://github.com/lookfirst | 1 |
| mattkanwisher | https://github.com/mattkanwisher | 1 |
| mKurrels | https://github.com/mKurrels | 1 |
| m1kola | https://github.com/m1kola | 1 |
| cavapoo2 | https://github.com/cavapoo2 | 1 |
| zeldrinn | https://github.com/zeldrinn | 1 |


