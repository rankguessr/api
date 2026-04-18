{
  pkgs,
  lib,
  config,
  inputs,
  ...
}:

{
  # required in .env: OSU_CLIENT_ID, OSU_CLIENT_SECRET, SENTRY_DSN
  dotenv.enable = true;

  env.PORT = "8080";
  env.WEB_URL = "http://localhost:5173";
  env.APP_URL = "http://localhost:8080";
  env.ENCRYPTION_KEY = "test-key-dont-use-in-production";

  # test turnstile secret, always returns success
  # use 2x0000000000000000000000000000000AA to fail
  env.TURNSTILE_SECRET = "1x0000000000000000000000000000000AA";
  env.DATABASE_URL = "postgres://postgres:postgres@127.0.0.1/rankguessr?sslmode=disable";
  env.REDIS_URL = "redis://127.0.0.1:6379";

  packages = [
    pkgs.git
    pkgs.air
  ];

  languages.go.enable = true;

  services.postgres = {
    enable = true;
    initialDatabases = [
      {
        name = "rankguessr";
        user = "postgres";
        pass = "postgres";
      }
    ];
    listen_addresses = "127.0.0.1";
    initialScript = ''
      CREATE ROLE postgres SUPERUSER;
    '';
  };

  services.redis = {
    enable = true;
    port = 6379;
    bind = "127.0.0.1";
  };

  processes = {
    backend = {
      exec = "go build -o guessr . && ./guessr start";
    };
  };
}
