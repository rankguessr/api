{
  pkgs,
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
  env.DATABASE_URL = "postgres://postgres:postgres@127.0.0.1:5432/rankguessr?sslmode=disable";
  env.REDIS_URL = "redis://127.0.0.1:6379";
  env.ENABLE_SENTRY = "0";

  packages = [
    pkgs.git
    pkgs.air
  ];

  languages.go.enable = true;

  scripts.buildcli.exec = "go build -o ./bin/guessr ./cmd/guessr";
  scripts.opendb.exec = "psql -U postgres -d rankguessr";

  services.redis = {
    enable = true;
    port = 6379;
    bind = "127.0.0.1";
  };

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

  processes = {
    backend = {
      exec = "buildcli && ./bin/guessr start --dev";
    };
  };
}
