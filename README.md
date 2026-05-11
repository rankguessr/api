[![GitHub License][license-shield]][license-url]
[![Go Report Card][go-report-shield]][go-report-url]
[![Website][website-shield]][website-url]
[![Discord][discord-shield]][discord-url]

# Rankguessr API
This repo contains all logic behind rankguessr and some tools for player pool collection

## Run Locally

Clone the project

```bash
  git clone https://github.com/rankguessr/api rankguessr-api
  cd rankguessr-api
```

Start the server (all the required services are started automatically)

```bash
  devenv up
```

## Environment Variables

To run this project locally, you will need to add the following environment variables to your .env file

`OSU_CLIENT_ID, OSU_CLIENT_SECRET` - can be acquired [here](https://osu.ppy.sh/home/account/edit)

Variables which are set in devenv.nix:

`PORT` - default is 8080

`WEB_URL, APP_URL` - web and api urls

`ENCRYPTION_KEY` - is used to encrypt client secrets in `sessions` table

`REDIS_URL, DATABASE_URL` - redis and postgres connection urls

`TURNSTILE_SECRET` - cloudflare's turnstile secret key, uses test key by default which will always return success

S3 variables:

`S3_SECURE, S3_ENDPOINT, S3_REGION, S3_BUCKET_NAME, S3_SECRET_KEY, S3_ACCESS_KEY`

## FAQ

#### How can i fill a `players` db table?

There is a built-in subcommand called `collect`, which for now only has one players source, which is top 10k players by country

Enter devenv shell:

```bash
devenv shell
```

Build rankguessr cli & start collecting players:

```bash
buildcli && ./bin/guessr collect top10k --country *country_code*
```

Default command output is `collected/top10k_*country_code*.csv`

Check `help` command for more options

[license-shield]: https://img.shields.io/github/license/rankguessr/api
[license-url]: https://github.com/rankguessr/api/blob/main/LICENSE
[website-shield]: https://img.shields.io/website?url=https%3A%2F%2Frankguessr.app
[website-url]: https://rankguessr.app
[discord-shield]: https://img.shields.io/discord/1495567491563261974
[discord-url]: https://discord.gg/WWFRdrrBfQ
[go-report-shield]: https://goreportcard.com/badge/github.com/rankguessr/api
[go-report-url]: https://goreportcard.com/report/github.com/rankguessr/api
