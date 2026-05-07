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

  env.S3_ENDPOINT = "127.0.0.1:9000";
  env.S3_REGION = "us-east-1";
  env.S3_BUCKET_NAME = "default";
  env.S3_PUBLIC_URL = "http://localhost:9000";
  env.S3_SECRET_KEY = "minioadmin";
  env.S3_ACCESS_KEY = "minioadmin";

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

  services.minio = {
    enable = true;
    afterStart = ''
      mc mb local/default
      mc anonymous set public local/default
    '';
  };

  processes = {
    backend = {
      exec = "buildcli && ./bin/guessr start --dev";
    };
  };
}
