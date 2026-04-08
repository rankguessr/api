{
  pkgs,
  lib,
  config,
  inputs,
  ...
}:

{
  dotenv.enable = true;

  env.WEB_URL = "http://localhost:5173";
  env.APP_URL = "http://localhost:8080";
  env.JWT_SECRET = "test-secret-dont-use-in-production";
  env.DATABASE_URL = "postgres://postgres:postgres@127.0.0.1/rankguessr?sslmode=disable";

  packages = [
    pkgs.git
    pkgs.air
  ];

  # git-hooks.hooks = {
  #   govet = {
  #     enable = true;
  #     pass_filenames = false;
  #   };
  #   golangci-lint = {
  #     enable = true;
  #     pass_filenames = false;
  #   };
  # };

  languages.go.enable = true;
  # languages.go.version = "1.25.5";

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
      exec = "air";
    };
  };
}
